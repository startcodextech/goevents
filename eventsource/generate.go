package eventsource

//go:generate mockery --quiet --name ".*(Aggregate|Repository|Store|Handler)$"  --inpackage --case underscore
