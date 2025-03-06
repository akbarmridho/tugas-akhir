-- Fine tuning Citus configuration
SET citus.shard_count = 64;

-- Add additional distribution key column
ALTER TABLE "orders" ADD COLUMN distribution_key bigint not null; -- should be filled with area id of the purchased ticket

-- Convert small, static tables to reference tables
SELECT create_reference_table('events');
SELECT create_reference_table('ticket_categories');
SELECT create_reference_table('ticket_sales');
SELECT create_reference_table('ticket_packages');
SELECT create_reference_table('ticket_areas');

-- Shard and colocate the rest of the tables
SELECT create_distributed_table('ticket_seats', 'ticket_area_id');
SELECT create_distributed_table('orders', 'distribution_key', colocate_with => 'ticket_seats');
SELECT create_distributed_table('order_items', 'order_id', colocate_with => 'orders');
SELECT create_distributed_table('invoices', 'order_id', colocate_with => 'orders');
SELECT create_distributed_table('issued_tickets', 'order_id', colocate_with => 'orders');
