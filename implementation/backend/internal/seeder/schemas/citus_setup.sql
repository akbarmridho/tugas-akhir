-- Drop distributed-distributed FK
ALTER TABLE order_items DROP CONSTRAINT order_items_ticket_seat_id_fk;

-- marker split
ALTER TABLE order_items DROP CONSTRAINT order_items_order_id_fk;

-- marker split
ALTER TABLE invoices DROP CONSTRAINT invoices_order_id_fk;

-- marker split
ALTER TABLE issued_tickets DROP CONSTRAINT issued_tickets_order_id_fk;

-- marker split
ALTER TABLE issued_tickets DROP CONSTRAINT issued_tickets_ticket_seat_id_fk;

-- marker split
ALTER TABLE issued_tickets DROP CONSTRAINT issued_tickets_order_items_id_fk;

-- marker split

-- Fine tuning Citus configuration
SET citus.shard_count = 32;
SET citus.max_adaptive_executor_pool_size = 2;
SET citus.max_cached_conns_per_worker = 4;
SET citus.executor_slow_start_interval = 75;

-- Convert small, static tables to reference tables
SELECT create_reference_table('events');
SELECT create_reference_table('ticket_categories');
SELECT create_reference_table('ticket_sales');
SELECT create_reference_table('ticket_packages');
SELECT create_reference_table('ticket_areas');

-- Distribute by ticket_area_id and colocate explicitly
SELECT create_distributed_table('ticket_seats', 'ticket_area_id');
SELECT create_distributed_table('orders', 'ticket_area_id', colocate_with => 'ticket_seats');
SELECT create_distributed_table('order_items', 'ticket_area_id', colocate_with => 'ticket_seats');
SELECT create_distributed_table('invoices', 'ticket_area_id', colocate_with => 'ticket_seats');
SELECT create_distributed_table('issued_tickets', 'ticket_area_id', colocate_with => 'ticket_seats');

-- marker split

CREATE INDEX IF NOT EXISTS idx_invoices_ticket_area_id_order_id ON invoices(ticket_area_id, order_id);

-- marker split
CREATE INDEX IF NOT EXISTS idx_order_items_ticket_area_id_order_id ON order_items(ticket_area_id, order_id);