# Run Notes - d2t4

Variant: stress-3
Scenario: sf-2
Flow Control: nofc
Database: postgres
Start Time: 2025-06-03 19:44 (WIB)
End Time: 2025-06-03 20:01 (WIB)

## Obversations

The test is good, but the current application does not automatically load balance the query so the primary PostgreSQL
is the only instance that are under heavy load.

Soo I need to find a way to implement the load balancing behaviour.

## Query Stats

If needed.
