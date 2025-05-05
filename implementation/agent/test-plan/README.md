# Test Plan

## Type of Tests

### Smoke Test

Test in local kubernetes cluster to check the validity of the system.

Category: smoke test.
Type: configure concurrent VUs.
Test duration: 5 minutes.
Scenario: scaled data.

### Real Simulation

User arrival follow lognormal distribution. The total of users and ticket are the scaled down version of the original calculation.

Category: spike test.
Type: configure arrival rate.
Test duration: 10 minutes.
Scenario: scaled data.
Variation: configure arrival rate and data scale.

### Race to 500k/ Stress Testing

Test for each system with constant UVs to see which one can serve 500k iterations faster. The number of iterations is large enough so that the test duration is long enough to act as stress testing at the same time.

Category: shared iteration tests (run until x iterations), stress testing.
Type: configure concurrent VUs.
Test duration: until finished.
Scenario: full data.
Variation: configure concurrent VUs.

### Breakpoint Testing

Spam as many VUs as possible until the system breaks. Increase the load gradually.

Category: breakpoint testing.
Type: configure concurrent VUs.
Test duration: until the system break or until not enough resources for the k6 agent.
Scenario: full data.
Variation: none.
