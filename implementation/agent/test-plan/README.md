# Test Plan

## Data Scenario

| Scenario Name | Day Count | Seat/ day | Total Seat |
| ------------- | --------- | --------- | ---------- |
| `sf-4`        | 4         | 80.000    | 320.000    |
| `sf-2`        | 2         | 80.000    | 160.000    |
| `sf-1`        | 1         | 80.000    | 80.000     |
| `s2-4`        | 4         | 40.000    | 160.000    |
| `s2-2`        | 2         | 40.000    | 80.000     |
| `s2-1`        | 1         | 40.000    | 40.000     |
| `s5-4`        | 4         | 16.000    | 64.000     |
| `s5-2`        | 2         | 16.000    | 32.000     |
| `s5-1`        | 1         | 16.000    | 16.000     |
| `s10-4`       | 4         | 8.000     | 32.000     |
| `s10-2`       | 2         | 8.000     | 16.000     |
| `s10-1`       | 1         | 8.000     | 8.000      |

## Type of Tests

### Smoke Test

Test in local kubernetes cluster to check the validity of the system.

Category: smoke test.
Type: configure concurrent VUs.
Test duration: 5 minutes.
Scenario: scaled data.

| Test # | Scenario Used | AVG VUs | Duration  |
| ------ | ------------- | ------- | --------- |
| 1      | `s10-2`       | 50      | 5 minutes |

### Real Simulation

User arrival follow lognormal distribution. The total of users and ticket are the scaled down version of the original calculation.

Category: spike test.
Type: configure arrival rate.
Test duration: 10 minutes.
Scenario: scaled data.
Variation: configure arrival rate and data scale.

| Test # | Scenario Used | Total Iter | Peak VUs | Duration   |
| ------ | ------------- | ---------- | -------- | ---------- |
| 1      | `s10-2`       | 40.000     | 15.000   | 10 minutes |

### Race to 350k/ Stress Testing

Test for each system with constant UVs to see which one can serve 350k iterations faster. The number of iterations is large enough so that the test duration is long enough to act as stress testing at the same time.

Category: shared iteration tests (run until x iterations), stress testing.
Type: configure concurrent VUs.
Test duration: until finished.
Scenario: full data.
Variation: configure concurrent VUs.

| Test # | Scenario Used | Total Iter | Constant UVs | Max Duration |
| ------ | ------------- | ---------- | ------------ | ------------ |
| 1      | `sf-4`        | 350.000    | 8.000        | 15 minutes   |
| 3      | `sf-4`        | 350.000    | 10.000       | 15 minutes   |
