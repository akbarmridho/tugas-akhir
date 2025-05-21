CREATE EXTENSION IF NOT EXISTS pg_prewarm;

ANALYZE events;
ANALYZE ticket_categories;
ANALYZE ticket_sales;
ANALYZE ticket_packages;
ANALYZE ticket_areas;
ANALYZE ticket_seats;

CHECKPOINT;

SELECT pg_sleep(10);

SELECT pg_prewarm('events');
SELECT pg_prewarm('ticket_categories');
SELECT pg_prewarm('ticket_sales');
SELECT pg_prewarm('ticket_packages');
SELECT pg_prewarm('ticket_areas');
SELECT pg_prewarm('ticket_seats');