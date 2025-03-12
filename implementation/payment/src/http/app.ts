import { redis } from "../common/redis.js";
import { logger } from "../common/logger.js";
import { OpenAPIHono, createRoute, z } from "@hono/zod-openapi";
import {
	CreateInvoiceSchema,
	ErrorSchema,
	IdParamsSchema,
	InvoiceSchema,
	PayInvoiceSchema,
	type InvoiceType,
} from "../common/schema.js";
import { HTTPException } from "hono/http-exception";
import { prometheus } from "@hono/prometheus";
import { logger as honoLogger } from "hono/logger";
import { generateId, instanceId } from "../common/utils.js";
import dayjs from "dayjs";
import { queue } from "./queue.js";
import { timeout } from "hono/timeout";

const { printMetrics, registerMetrics } = prometheus({
	metricOptions: {
		requestDuration: {
			customLabels: {
				instanceId: () => instanceId,
			},
		},
		requestsTotal: {
			customLabels: {
				instanceId: () => instanceId,
			},
		},
	},
});

export const app = new OpenAPIHono();

app.use("*", timeout(30000));
app.use("/invoices/*", registerMetrics);

app.use(
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
	const metrics = await queue.exportPrometheusMetrics();
	return c.text(metrics);
});

// healthcheck endpoint
app.get("/health", async (c) => {
	try {
		logger.info("before ping");
		const retrievedValue = await redis.ping();
		logger.info("after ping");

		if (retrievedValue === "PONG") {
			return c.json({
				status: "healthy",
				message: "Node healthy",
			});
		}

		return c.json(
			{
				status: "unhealthy",
				message: "Redis cluster returned unexpected response",
			},
			500,
		);
	} catch (error) {
		logger.error("Redis health check failed", { error });
		return c.json(
			{
				status: "unhealthy",
				message: "Redis cluster connection failed",
			},
			503,
		);
	}
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

		const data = await redis.get(`invoices:${id}`);

		if (data === null) {
			return c.json(
				{
					message: "Invoice not found",
				},
				404,
			);
		}

		const invoice = await InvoiceSchema.safeParseAsync(JSON.parse(data));

		if (!invoice.success) {
			throw new HTTPException(500, {
				message: "Cannot parse invoice data from redis",
			});
		}

		return c.json(invoice.data, 200);
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
		const id = generateId();

		const payload = c.req.valid("json");

		const now = dayjs();
		const expireDate = now.add(10, "minute");

		const data: InvoiceType = {
			id,
			amount: payload.amount,
			description: payload.description,
			externalId: payload.externalId,
			createdAt: now.toDate(),
			expiredAt: expireDate.toDate(),
			paidAt: null,
			paidAmount: null,
			status: "pending",
		};

		await redis.setex(
			`invoices:${id}`,
			5 * 60 * 60 * 1000,
			JSON.stringify(data),
		);

		await queue.add(
			"webhook",
			{ id },
			{
				jobId: id,
				delay: expireDate.diff(now) + 1000,
			},
		);

		return c.json(data, 200);
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

		const data = await redis.get(`invoices:${id}`);

		if (data === null) {
			return c.json(
				{
					message: "Invoice not found",
				},
				404,
			);
		}

		const rawInvoice = await InvoiceSchema.safeParseAsync(JSON.parse(data));

		if (!rawInvoice.success) {
			throw new HTTPException(500, {
				message: "Cannot parse invoice data from redis",
			});
		}

		const invoice = rawInvoice.data;

		if (
			invoice.status !== "pending" ||
			dayjs(invoice.expiredAt).isBefore(dayjs())
		) {
			return c.json(
				{
					message: "Invoice status must be pending",
				},
				400,
			);
		}

		const payload = c.req.valid("json");

		if (payload.mode === "success") {
			invoice.paidAmount = invoice.amount;
			invoice.paidAt = new Date();
			invoice.status = "paid";
		} else {
			invoice.status = "failed";
		}

		await redis.setex(
			`invoices:${id}`,
			5 * 60 * 60 * 1000,
			JSON.stringify(invoice),
		);

		const job = await queue.getJob(id);

		if (job && (await job.isDelayed())) {
			await job.promote();
		} else {
			logger.warn(`Cannot find corresponding job for invoice ${id}`);
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
