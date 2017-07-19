package chargeback

// Create Con***REMOVED***gMap with buckets, paths, and secrets for AWS data, promsum data, and query output
// Check that adequate credentials exist for S3 data of promsum and AWS billing
// Check that Hive cluster has been setup, if not deploy
// Check that Presto cluster has been setup, if not deploy

// Determine current assembly ID by reading the *-Manifest.json from AWS billing data
// Create a hive table for billing data using queries/setup-aws.hive substituting the Assembly ID from the step above
// Use promsum to collect usage information for the given range
// Create a hive table for promsum data using queries/setup-promsum.hive
// Determine cost per pod by running the query cost-per-pod.presto
// Write results back to S3 bucket
