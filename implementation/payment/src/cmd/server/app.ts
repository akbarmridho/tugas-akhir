import { OpenAPIHono, createRoute } from "@hono/zod-openapi";
import {
	CreateInvoiceSchema,
	InvoiceSchema,
	PayInvoiceSchema,
} from "../../entity/invoice.js";
import { HTTPException } from "hono/http-exception";
import { prometheus } from "@hono/prometheus";
import { logger as honoLogger } from "hono/logger";
import dayjs from "dayjs";
import type { Logger } from "winston";
import { timeout } from "hono/timeout";
import { IdGenerator } from "../../utils/id-generator.js";
import type { WebhookQueue } from "../../infrastructure/queue.js";
import { HealthService } from "../../service/health.js";
import { ErrorSchema, IdParamsSchema } from "../../entity/http.js";
import { InvoiceService } from "../../service/invoice.js";
import type { Cluster } from "ioredis";

export const createApp = (
	logger: Logger,
	redis: Cluster,
	webhookQueue: WebhookQueue,
) => {
	const idGenerator = new IdGenerator();
	const healthService = new HealthService(redis, logger);
	const invoiceService = new InvoiceService(redis, logger, idGenerator);

	const { printMetrics, registerMetrics } = prometheus({
		prefix: "payment",
	});

	const app = new OpenAPIHono();

	app.use("*", timeout(30000));
	app.use("/invoices/*", registerMetrics);

	app.use(
		"/invoices/*",
		honoLogger((message, meta) => {
			logger.info(message, meta);
		}),
	);

	app.notFound((c) => c.json({ message: "Not Found", ok: false }, 404));

	app.onError((err, c) => {
		const status = err instanceof HTTPException ? err.status : 500;
		logger.error(`${c.req.method} ${c.req.path}`, {
			error: err,
			status,
			message: err.message,
		});
		c.status(status);
		return c.json({ message: err.message });
	});

	app.get("/", (c) => {
		return c.text("Payment Service");
	});

	app.get("/metrics", printMetrics);

	app.get("/metrics-queue", async (c) => {
		const metrics = await webhookQueue.exportPrometheusMetrics();
		return c.text(metrics);
	});

	// healthcheck endpoint
	app.get("/health", async (c) => {
		const status = await healthService.check();
		return c.json(status.data, status.code);
	});

	// handle get invoice inffo
	app.openapi(
		createRoute({
			method: "get",
			summary: "Get Invoice",
			path: "/invoices/{id}",
			request: {
				params: IdParamsSchema,
			},
			responses: {
				200: {
					content: {
						"application/json": {
							schema: InvoiceSchema,
						},
					},
					description: "Retrieve the invoice",
				},
				404: {
					content: {
						"application/json": {
							schema: ErrorSchema,
						},
					},
					description: "Invoice not found",
				},
			},
		}),
		async (c) => {
			const id = c.req.valid("param").id;

			const invoice = await invoiceService.getInvoice(id);

			if (invoice === null) {
				return c.json(
					{
						message: "Invoice not found",
					},
					404,
				);
			}

			return c.json(invoice, 200);
		},
	);

	// handle create payment
	app.openapi(
		createRoute({
			method: "post",
			summary: "Create Invoice",
			path: "/invoices",
			request: {
				body: {
					content: {
						"application/json": {
							schema: CreateInvoiceSchema,
						},
					},
				},
			},
			responses: {
				200: {
					content: {
						"application/json": {
							schema: InvoiceSchema,
						},
					},
					description: "Retrieve the invoice",
				},
			},
		}),
		async (c) => {
			const payload = c.req.valid("json");

			const invoice = await invoiceService.createInvoice(payload);

			await webhookQueue.add(
				"webhook",
				{ id: invoice.id },
				{
					jobId: invoice.id,
					delay: dayjs(invoice.expiredAt).diff(dayjs()) + 1000,
				},
			);

			return c.json(invoice, 200);
		},
	);

	// handle invoice payment
	app.openapi(
		createRoute({
			method: "post",
			summary: "Pay Invoice",
			path: "/invoices/{id}/payment",
			request: {
				params: IdParamsSchema,
				body: {
					content: {
						"application/json": {
							schema: PayInvoiceSchema,
						},
					},
				},
			},
			responses: {
				200: {
					content: {
						"application/json": {
							schema: InvoiceSchema,
						},
					},
					description: "Retrieve the invoice",
				},
				400: {
					content: {
						"application/json": {
							schema: ErrorSchema,
						},
					},
					description: "Bad request",
				},
				404: {
					content: {
						"application/json": {
							schema: ErrorSchema,
						},
					},
					description: "Invoice not found",
				},
			},
		}),
		async (c) => {
			const id = c.req.valid("param").id;
			const payload = c.req.valid("json");

			const invoice = await invoiceService.payInvoice(id, payload);

			if (invoice === null) {
				return c.json(
					{
						message: "Invoice not found",
					},
					404,
				);
			}

			const job = await webhookQueue.getJob(invoice.id);

			if (job && (await job.isDelayed())) {
				await job.promote();
			} else {
				logger.warn(`Cannot find corresponding job for invoice ${invoice.id}`);
			}

			return c.json(invoice, 200);
		},
	);

	// The OpenAPI documentation will be available at /doc
	app.doc("/doc", {
		openapi: "3.0.0",
		info: {
			version: "1.0.0",
			title: "Payment Service",
		},
	});

	return app;
};
