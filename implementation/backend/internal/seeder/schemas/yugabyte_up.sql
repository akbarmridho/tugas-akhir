CREATE TYPE area_type AS ENUM ('numbered-seating', 'free-standing');
CREATE TYPE seat_status AS ENUM ('available', 'on-hold', 'sold');
CREATE TYPE order_status AS ENUM ('waiting-for-payment', 'failed', 'success');
CREATE TYPE invoice_status AS ENUM ('pending', 'expired', 'failed', 'paid');

CREATE TABLE "events" (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    location text NOT NULL,
    description text NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE "ticket_categories" (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    event_id bigint NOT NULL REFERENCES events(id),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE "ticket_sales" (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    sale_begin_at timestamptz NOT NULL,
    sale_end_at timestamptz NOT NULL,

    event_id bigint NOT NULL REFERENCES events(id),

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE "ticket_packages" (
    id bigserial PRIMARY KEY,
    price int NOT NULL,

    ticket_category_id bigint NOT NULL REFERENCES ticket_categories(id),
    ticket_sale_id bigint NOT NULL REFERENCES ticket_sales(id),

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE "ticket_areas" (
    id bigserial PRIMARY KEY,
    type area_type NOT NULL,

    ticket_package_id bigint NOT NULL REFERENCES ticket_packages(id),

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE "ticket_seats" (
    ticket_area_id bigint NOT NULL REFERENCES ticket_areas(id),
    id bigserial,
    seat_number text NOT NULL,
    status seat_status NOT NULL DEFAULT 'available',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (ticket_area_id HASH, id)
);

CREATE TABLE "orders" (
    ticket_area_id bigint NOT NULL REFERENCES ticket_areas(id),
    id bigserial,
    status order_status NOT NULL,
    fail_reason text,
    event_id bigint NOT NULL REFERENCES events(id),
    ticket_sale_id bigint NOT NULL REFERENCES ticket_sales(id),
    external_user_id text NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (ticket_area_id HASH, id)
);

CREATE TABLE "order_items" (
    ticket_area_id bigint NOT NULL,
    id bigserial,
    customer_name text NOT NULL,
    customer_email text NOT NULL,
    price int NOT NULL,
    order_id bigint NOT NULL,
    ticket_category_id bigint NOT NULL REFERENCES ticket_categories(id),
    ticket_seat_id bigint NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (ticket_area_id HASH, id),
    CONSTRAINT fk_order_items_orders FOREIGN KEY (ticket_area_id, order_id) REFERENCES orders(ticket_area_id, id),
    CONSTRAINT fk_order_items_ticket_seats FOREIGN KEY (ticket_area_id, ticket_seat_id) REFERENCES ticket_seats(ticket_area_id, id)
);

CREATE TABLE "invoices" (
    ticket_area_id bigint NOT NULL,
    id bigserial,
    status invoice_status NOT NULL DEFAULT 'pending',
    amount int NOT NULL,
    external_id text NOT NULL,
    order_id bigint NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (ticket_area_id HASH, id),
    CONSTRAINT fk_invoices_orders FOREIGN KEY (ticket_area_id, order_id) REFERENCES orders(ticket_area_id, id)
);

CREATE TABLE "issued_tickets" (
    ticket_area_id bigint NOT NULL,
    id bigserial,
    serial_number text NOT NULL,
    holder_name text NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    ticket_seat_id bigint NOT NULL,
    order_id bigint NOT NULL,
    order_item_id bigint NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (ticket_area_id HASH, id),
    CONSTRAINT fk_issued_tickets_orders FOREIGN KEY (ticket_area_id, order_id) REFERENCES orders(ticket_area_id, id),
    CONSTRAINT fk_issued_tickets_ticket_seats FOREIGN KEY (ticket_area_id, ticket_seat_id) REFERENCES ticket_seats(ticket_area_id, id),
    CONSTRAINT fk_issued_tickets_order_items FOREIGN KEY (ticket_area_id, order_item_id) REFERENCES order_items(ticket_area_id, id)
);

CREATE INDEX idx_ticket_seats_ticket_area_id_seat_number ON ticket_seats(ticket_area_id, seat_number);

CREATE INDEX idx_ticket_seats_ticket_area_id ON ticket_seats(ticket_area_id);

CREATE INDEX idx_ticket_packages_ticket_sale_id ON ticket_packages(ticket_sale_id);

CREATE INDEX idx_ticket_areas_ticket_package_id ON ticket_areas(ticket_package_id);
