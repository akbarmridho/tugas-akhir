
-- Setup CDC
CREATE SOURCE pg_source WITH (
    connector='postgres-cdc',
    hostname='<PG_HOST>',
    port='<PG_PORT>',
    username='<PG_USERNAME>',
    password='<PG_PASSWORD>',
    database.name='<PG_DATABASE>'
);

CREATE TABLE ticket_packages_rw
(
    id bigserial primary key,
    price int not null,
    ticket_category_id bigint not null,
    ticket_sale_id bigint not null,
    created_at timestamptz,
    updated_at timestamptz
)
FROM pg_source TABLE 'public.ticket_packages';

CREATE TABLE ticket_areas_rw (
     id bigserial primary key,
     type varchar not null,
     ticket_package_id  bigint not null,
     created_at timestamptz default now(),
     updated_at timestamptz default now()
)
FROM pg_source TABLE 'public.ticket_areas';

CREATE TABLE ticket_seats_rw (
    id bigserial primary key,
    seat_number text not null,
    status varchar not null,
    ticket_area_id bigint not null,
    created_at timestamptz,
    updated_at timestamptz
) WITH (
    connector = 'citus-cdc',
    hostname = '<PG_COORDINATOR_HOST>',
    port = '<PG_COORDINATOR_PORT>',
    username = '<PG_USERNAME>',
    password = '<PG_PASSWORD>',
    database.servers = '<PG_WORKERS_HOST>',
    database.name = '<PG_DATABASE>',
    table.name = 'ticket_seats'
);

CREATE MATERIALIZED VIEW ticket_availability AS
SELECT
    tp.ticket_sale_id AS ticket_sale_id,
    tp.id AS ticket_package_id,
    ta.id AS ticket_area_id,
    COUNT(ts.id) AS total_seats,
    COUNT(CASE WHEN ts.status = 'available' THEN 1 END) AS available_seats
FROM
    ticket_packages_rw tp
    INNER JOIN
    ticket_areas_rw ta ON ta.ticket_package_id = tp.id
    INNER JOIN
    ticket_seats_rw ts ON ts.ticket_area_id = ta.id
GROUP BY
    tp.id, ta.id, tp.ticket_sale_id;

CREATE INDEX ticket_availability_by_sale_id ON ticket_availability(ticket_sale_id);