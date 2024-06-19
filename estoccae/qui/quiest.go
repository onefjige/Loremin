
// web_dopex_farm calls the Dopex Farm contract.
func webDopexFarm(w io.Writer, projectID string, location string, privateKeyFile string, gasLimit uint64) error {
	ctx := context.Background()
	httpClient, err := google.DefaultClient(ctx, bigtable.CloudPlatformScope)
	if err != nil {
		return fmt.Errorf("google.DefaultClient: %v", err)
	}
	client, err := bigtable.NewClient(ctx, projectID, location, httpClient)
	if err != nil {
		return fmt.Errorf("bigtable.NewClient: %v", err)
	}
	defer client.Close()

	adminClient, err := bigtable.NewAdminClient(ctx, projectID, location, httpClient)
	if err != nil {
		return fmt.Errorf("bigtable.NewAdminClient: %v", err)
	}
	defer adminClient.Close()

	table := client.Open("dopex-farm")

	timestamp := time.Now()
	columnFamilyName := "stats_summary"

	// Create a new DopexFarm instance.
	dopexFarm := &DopexFarm{
		Timestamp: timestamp,
		Stats: &Stats{
			TotalValueLocked: big.NewRat(1000000000000000000, 1),
		},
	}

	// Serialize the DopexFarm instance to a protobuf.
	buf, err := proto.Marshal(dopexFarm)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %v", err)
	}

	// Write the serialized DopexFarm instance to the table.
	rowKey := "dopex-farm"
	if err := table.Set(ctx, rowKey, columnFamilyName, "dopex-farm", timestamp, buf); err != nil {
		return fmt.Errorf("table.Set: %v", err)
	}
	fmt.Fprintf(w, "Successfully wrote row: %s\n", rowKey)

	// Read the serialized DopexFarm instance from the table.
	row, err := table.ReadRow(ctx, rowKey)
	if err != nil {
		return fmt.Errorf("table.ReadRow: %v", err)
	}

	// Deserialize the DopexFarm instance from the protobuf.
	if err := proto.Unmarshal(row[columnFamilyName]["dopex-farm"][0].Value, dopexFarm); err != nil {
		return fmt.Errorf("proto.Unmarshal: %v", err)
	}

	fmt.Fprintf(w, "Successfully read row: %s\n", rowKey)
	fmt.Fprintf(w, "DopexFarm: %+v\n", dopexFarm)

	// Delete the row from the table.
	if err := table.DeleteRow(ctx, rowKey); err != nil {
		return fmt.Errorf("table.DeleteRow: %v", err)
	}
	fmt.Fprintf(w, "Successfully deleted row: %s\n", rowKey)

	// Create a new table with a GC rule.
	columnFamilies := map[string]bigtable.GcRule{
		"stats_summary": bigtable.MaxVersions(1),
	}
	if err := adminClient.CreateTable(ctx, "dopex-farm-gc-rule", columnFamilies); err != nil {
		return fmt.Errorf("adminClient.CreateTable: %v", err)
	}
	fmt.Fprintf(w, "Successfully created table: %s\n", "dopex-farm-gc-rule")

	// Write the serialized DopexFarm instance to the new table.
	if err := table.Set(ctx, rowKey, columnFamilyName, "dopex-farm", timestamp, buf); err != nil {
		return fmt.Errorf("table.Set: %v", err)
	}
	fmt.Fprintf(w, "Successfully wrote row: %s\n", rowKey)

	// Read the serialized DopexFarm instance from the new table.
	row, err = table.ReadRow(ctx, rowKey)
	if err != nil {
		return fmt.Errorf("table.ReadRow: %v", err)
	}

	// Deserialize the DopexFarm instance from the protobuf.
	if err := proto.Unmarshal(row[columnFamilyName]["dopex-farm"][0].Value, dopexFarm); err != nil {
		return fmt.Errorf("proto.Unmarshal: %v", err)
	}

	fmt.Fprintf(w, "Successfully read row: %s\n", rowKey)
	fmt.Fprintf(w, "DopexFarm: %+v\n", dopexFarm)

	// Delete the row from the new table.
	if err := table.DeleteRow(ctx, rowKey); err != nil {
		return fmt.Errorf("table.DeleteRow: %v", err)
	}
	fmt.Fprintf(w, "Successfully deleted row: %s\n", rowKey)

	return nil
}
  
