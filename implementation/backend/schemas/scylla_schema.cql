CREATE KEYSPACE IF NOT EXISTS ticket_system WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3};

-- Create User Defined Types for nested data
CREATE TYPE ticket_system.order_item_type (
    id bigint,
    customer_name text,
    customer_email text,
    price int,
    ticket_category_id bigint,
    ticket_seat_id bigint,
    created_at timestamp,
    updated_at timestamp
);

-- Single invoice as direct columns
CREATE TABLE ticket_system.orders (
    id bigint,
    status text, -- 'waiting-for-payment', 'failed', 'success'
    fail_reason text,
    event_id bigint,
    ticket_sale_id bigint,
    first_ticket_area_id bigint,
    external_user_id text,
    created_at timestamp,
    updated_at timestamp,

    -- Denormalized order items as a list of UDTs
    items list<frozen<order_item_type>>,

    -- Single invoice fields (previously in separate table)
    invoice_id bigint,
    invoice_status text, -- 'pending', 'expired', 'failed', 'paid'
    invoice_amount int,
    invoice_external_id text,
    invoice_created_at timestamp,
    invoice_updated_at timestamp,

    PRIMARY KEY (id)
);

-- Ticket seats tables remain unchanged
CREATE TABLE ticket_system.ticket_seats_numbered (
    id bigint,
    seat_number text,
    status text,
    ticket_area_id bigint,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (id)
);

CREATE TABLE ticket_system.ticket_seats_area (
    ticket_area_id bigint,
    id bigint,
    seat_number text,
    status text,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (ticket_area_id, id)
) WITH CLUSTERING ORDER BY (id ASC);

CREATE INDEX ON ticket_system.ticket_seats_area(id);
CREATE INDEX ON ticket_system.orders(invoice_id);