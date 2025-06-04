# Agent

## Prerequisites

- Download `k6` from [here](https://k6.io/docs/get-started/installation/).
- Install `xk6` `go install go.k6.io/xk6/cmd/xk6@latest`.
- Install dependencies `pnpm install`.

List of k6 extensions used:

- [xk6-faker](https://github.com/grafana/xk6-faker)

From the current folder, build k6 with extensions:

```bash
xk6 build --with github.com/grafana/xk6-faker@latest
```

## Agent Behavior

### Agent Flow

- Call the get all events API (`ticket/eventRoutesGetEvents`).
- Select the first event. Call get event API (`ticket/eventRoutesGetEvent`).
- Select ticket sale. Call the get availability API (`ticket/eventRoutesGetAvailability`).
- Select a ticket area. Call the get seats API (`ticket/eventRoutesGetSeats`).
- Place the order. Call the place order API (`ticket/orderRoutesPlaceOrder`).
- Paid the invoice. Call the paid invoice API (`payment/postInvoicesIdPayment`).
- Wait for x seconds. Call the get order API (`ticket/orderRoutesGetOrder`).
- If succeed, check the published issued tickets. Call the get issued tickets API (`ticket/orderRoutesGetIssuedTickets`).

### Agent Variance

- **Day Preference**
  - `specific-day` - can be one specific day or some specific day.
  - `any-day` - focus on target category availability rather than day selection.
- **Seating-Tier Preference**
  - `seated-high-tier` - focus on ticket category `platinum` and `gold` tier.
  - `seated-mid-tier` - focus on ticket category `silver`, `gold`, and `bronze` tier.
  - `seated-low-tier` - focus on ticket category `bronze` and `silver`.
  - `area-high-tier` - focus on ticket category `VIP` and `Zone A` tier.
  - `area-mid-tier` - focus on ticket category `Zone B` and `Zone A` tier.
- **Ticket Quantity Preference**
  - `solo` - purchase 1 ticket.
  - `couple` - purchase 2 tickets.
  - `group` - purchase 3-5 tickets.
- **Persistence**
  - `high` - iterate until the preferred ticket is purchased with at most 27 browse attempts and 9 order attempts.
  - `medium` - iterate until the preferred ticket is purchased with at most 18 browse attempts and 6 order attempts.
  - `low` - iterate until the purchased ticket is purchased with at most 9 browse attempts and 3 order attempts.

### Profile Distribution

**Payment Failure:**

When the user successfully placed an order, there are 5% probability that the user's payment fails for simulating user backing off or cancelling the payment.

**Seating Distribution:**

| Tier        | Distribution |
| ----------- | ------------ |
| Seated Low  | 27%          |
| Seated Mid  | 26%          |
| Seated High | 19%          |
| Area Mid    | 17%          |
| Area High   | 11%          |

**Day Distribution:**

| Day      | Distribution |
| -------- | ------------ |
| Specific | 52%          |
| Any      | 48%          |

**Persistence:**

| Persistence | Distribution |
| ----------- | ------------ |
| Low         | 19%          |
| Medium      | 51%          |
| High        | 30%          |

**Quantity:**

| Quantity | Distribution |
| -------- | ------------ |
| Solo     | 28%          |
| Couple   | 48%          |
| Group    | 24%          |

**Profile Distribution:**

| Day Preference | Seating Tier | Quantity | Persistence | Percentage |
| -------------- | ------------ | -------- | ----------- | ---------- |
| Specific       | Seated-Low   | Solo     | Medium      | 1.00%      |
| Specific       | Seated-Low   | Solo     | High        | 2.00%      |
| Specific       | Seated-Low   | Couple   | Low         | 2.00%      |
| Specific       | Seated-Low   | Couple   | Medium      | 3.00%      |
| Specific       | Seated-Low   | Couple   | High        | 2.00%      |
| Specific       | Seated-Low   | Group    | Low         | 2.00%      |
| Specific       | Seated-Low   | Group    | Medium      | 2.00%      |
| Specific       | Seated-Mid   | Solo     | Medium      | 1.00%      |
| Specific       | Seated-Mid   | Solo     | High        | 2.00%      |
| Specific       | Seated-Mid   | Couple   | Low         | 2.00%      |
| Specific       | Seated-Mid   | Couple   | Medium      | 2.00%      |
| Specific       | Seated-Mid   | Couple   | High        | 2.00%      |
| Specific       | Seated-Mid   | Group    | Low         | 3.00%      |
| Specific       | Seated-Mid   | Group    | Medium      | 2.00%      |
| Specific       | Seated-High  | Solo     | Medium      | 2.00%      |
| Specific       | Seated-High  | Solo     | High        | 2.00%      |
| Specific       | Seated-High  | Couple   | Medium      | 2.00%      |
| Specific       | Seated-High  | Couple   | High        | 2.00%      |
| Specific       | Seated-High  | Group    | Medium      | 2.00%      |
| Specific       | Area-Mid     | Solo     | Low         | 1.00%      |
| Specific       | Area-Mid     | Solo     | Medium      | 2.00%      |
| Specific       | Area-Mid     | Couple   | Low         | 1.00%      |
| Specific       | Area-Mid     | Couple   | Medium      | 2.00%      |
| Specific       | Area-Mid     | Couple   | High        | 1.00%      |
| Specific       | Area-Mid     | Group    | Medium      | 2.00%      |
| Specific       | Area-High    | Solo     | Medium      | 2.00%      |
| Specific       | Area-High    | Solo     | High        | 1.00%      |
| Specific       | Area-High    | Couple   | Medium      | 1.00%      |
| Specific       | Area-High    | Couple   | High        | 1.00%      |
| Any Day        | Seated-Low   | Solo     | Medium      | 2.00%      |
| Any Day        | Seated-Low   | Couple   | Low         | 3.00%      |
| Any Day        | Seated-Low   | Couple   | Medium      | 3.00%      |
| Any Day        | Seated-Low   | Couple   | High        | 3.00%      |
| Any Day        | Seated-Low   | Group    | Medium      | 2.00%      |
| Any Day        | Seated-Mid   | Solo     | Medium      | 2.00%      |
| Any Day        | Seated-Mid   | Couple   | Low         | 3.00%      |
| Any Day        | Seated-Mid   | Couple   | Medium      | 1.00%      |
| Any Day        | Seated-Mid   | Couple   | High        | 2.00%      |
| Any Day        | Seated-Mid   | Group    | Medium      | 4.00%      |
| Any Day        | Seated-High  | Solo     | Medium      | 2.00%      |
| Any Day        | Seated-High  | Solo     | High        | 2.00%      |
| Any Day        | Seated-High  | Group    | Medium      | 3.00%      |
| Any Day        | Seated-High  | Couple   | High        | 2.00%      |
| Any Day        | Area-Mid     | Solo     | Medium      | 2.00%      |
| Any Day        | Area-Mid     | Couple   | Low         | 2.00%      |
| Any Day        | Area-Mid     | Couple   | High        | 2.00%      |
| Any Day        | Area-Mid     | Group    | Medium      | 2.00%      |
| Any Day        | Area-High    | Solo     | High        | 2.00%      |
| Any Day        | Area-High    | Couple   | Medium      | 2.00%      |
| Any Day        | Area-High    | Couple   | High        | 2.00%      |

## Tests

Refer to [this docs](./test-plan/README.md) for test plan.

## Running Tests

Required envs:

- `RUN_ID` any randomly generated unique string.
- `VARIANT` with the following values: `debug`, `smoke`, `smokey`, `sim-1`, `sim-2`, `sim-test`, `stress-1`, `stress-2`. This differentiate the k6 agent request pattern and behaviour.

### Running Locally

Generate random run ID.

```bash
openssl rand -hex 6
```

Prepare the environment.

```bash
export RUN_ID=<any string>
export VARIANT=<variant>
```

For example

```bash
export RUN_ID=906023a363ed
export VARIANT=debug
```

Running the code.

```bash
npm run test
# or
npm run test-debug
```
