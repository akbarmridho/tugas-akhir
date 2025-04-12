CREATE TYPE area_type as ENUM ('numbered-seating', 'free-standing');
CREATE TYPE seat_status as ENUM ('available', 'on-hold', 'sold');
CREATE TYPE order_status as ENUM ('waiting-for-payment', 'failed', 'success');
CREATE TYPE invoice_status as ENUM ('pending', 'expired', 'failed', 'paid');

CREATE TABLE "events" (
    id bigserial primary key,
    name text not null,
    location text not null,
    description text not null,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "ticket_categories" (
    id bigserial primary key,
    name text not null,

    event_id bigint not null references events(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "ticket_sales" (
    id bigserial primary key,
    name text not null,
    sale_begin_at timestamptz not null,
    sale_end_at timestamptz not null,

    event_id bigint not null references events(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "ticket_packages" (
    id bigserial primary key,
    price int not null,

    ticket_category_id bigint not null references ticket_categories(id) on update cascade on delete cascade,
    ticket_sale_id bigint not null references ticket_sales(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "ticket_areas" (
    id bigserial primary key,
    type area_type not null,

    ticket_package_id  bigint not null references ticket_packages(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "ticket_seats" (
    id bigserial primary key,
    seat_number text not null,
    status seat_status not null default 'available',

    ticket_area_id bigint not null references ticket_areas(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "orders" (
    id bigserial primary key,
    status order_status not null,
    fail_reason text,

    event_id bigint not null references events(id) on update cascade on delete cascade,
    ticket_sale_id bigint not null references ticket_sales(id) on update cascade on delete cascade,
    first_ticket_area_id bigint not null references ticket_areas(id) on update cascade on delete cascade,

    external_user_id text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "order_items" (
    id bigserial primary key,
    customer_name text not null,
    customer_email text not null,
    price int not null,

    order_id bigint not null references orders(id) on update cascade on delete cascade,
    ticket_category_id bigint not null references ticket_categories(id) on update cascade on delete cascade,
    ticket_seat_id bigint not null references ticket_seats(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "invoices"
(
    id     bigserial primary key,
    status invoice_status not null default 'pending',
    amount int not null,
    external_id text not null,

    order_id bigint not null references orders(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE "issued_tickets" (
    id bigserial primary key,
    serial_number text not null,
    holder_name text not null,
    name text not null,
    description text not null,

    seat_id bigint not null references ticket_seats(id) on update cascade on delete cascade,
    order_id bigint not null references orders(id) on update cascade on delete cascade,
    order_item_id bigint not null references order_items(id) on update cascade on delete cascade,

    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE INDEX idx_ticket_seats_ticket_area_id_seat_number ON ticket_seats(ticket_area_id, seat_number);

CREATE INDEX idx_ticket_seats_ticket_area_id ON ticket_seats(ticket_area_id);

CREATE INDEX idx_ticket_packages_ticket_sale_id ON ticket_packages(ticket_sale_id);

CREATE INDEX idx_ticket_areas_ticket_package_id ON ticket_areas(ticket_package_id);