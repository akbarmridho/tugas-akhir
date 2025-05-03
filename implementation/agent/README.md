# Agent

## Prerequisites

- Download `k6` from [here](https://k6.io/docs/get-started/installation/).
- Install `xk6` `go install go.k6.io/xk6/cmd/xk6@latest`.
- Install dependencies `pnpm install`.

List of k6 extensions used:

- [xk6-faker](https://github.com/grafana/xk6-faker)

Build k6 with extensions:

```bash
xk6 build --with github.com/grafana/xk6-faker@latest
```

### Installing xk6-dashboard with Docker

[xk6-dashboard](https://github.com/grafana/xk6-dashboard) is a k6 extension that can be used to visualise your performance test in real time.

To run the tests with monitoring with xk6-dashboard extension, we need to install it. The simplest way to install is via docker and can be done via

`docker pull ghcr.io/grafana/xk6-dashboard:0.6.1`

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

### reqres

We use the [reqres](https://reqres.in/) publicly hosted REST API to showcase the testing with k6

To execute the first sample test that showcases how `per-vu-iterations` works, you can run:

`yarn test:demo`

To test with monitoring in place, run:

`yarn test-with-monitoring:demo`

To execute the second sample test that showcases how to use `stages`, you can run:

`yarn test:demo-stages`

To test with monitoring in place, run:

`yarn test-with-monitoring:demo-stages`
