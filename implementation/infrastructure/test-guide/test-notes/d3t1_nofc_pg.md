# Run Notes - d3t1

Variant: stress-3
Scenario: sf-4
Flow Control: nofc
Database: postgres
Start Time: 2025-06-04 18:56 (WIB)
End Time: 2025-06-04 19:13 (WIB)

## Obversations

Awal-awal oke ketika masih banyak tiket yang bisa dijual. Request naik tajam dan Redis gak bisa handle ketika kebanyakan tiket sudah terjual.

PGCat pakai resource lebih banyak dari perkiraan (2vCPU).

Result: kurangi satu instance ticket dan alihkan 2vCPU ke PGCat.

Test diulang tapi dengan 10k VU.

Prometheus data not saved.

## Query Stats

If needed.
