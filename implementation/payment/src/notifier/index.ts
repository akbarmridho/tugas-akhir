import { Worker } from "bullmq";
import { logger } from "../common/logger.js";
import { redis } from "../common/redis.js";
import { env } from "../common/env.js";
import { InvoiceSchema } from "../common/schema.js";
import got from "got";
import { readFileSync } from "node:fs";

(async function main() {
	// Create the worker to process webhook jobs
	const worker = new Worker<{ id: string }, void, "webhook">(
		"{webhook}",
		async (job) => {
			const { id } = job.data;

			const l = logger.child({ id });
			l.info("handling job");

			const data = await redis.get(`invoices:${id}`);

			if (data === null) {
				throw new Error("Invoice Not Found");
			}

			const rawInvoice = await InvoiceSchema.safeParseAsync(JSON.parse(data));

			if (!rawInvoice.success) {
				throw new Error("Cannot parse invoice data from redis");
			}

			const invoice = rawInvoice.data;

			try {
				const response = await got.post(env.WEBHOOK_URL, {
					json: invoice,
					https: {
						key: readFileSync("../cert/key.pem"),
						certificate: readFileSync("../cert/cert.pem"),
						rejectUnauthorized: false,
					},
					http2: true,
				});

				if (response.ok) {
					l.info(`Webhook success ${id}`);
				} else if (response.statusCode >= 400) {
					const statusCode = response.statusCode;

					let errorBody: string;
					try {
						const body = JSON.parse(response.body);

						// biome-ignore lint/suspicious/noExplicitAny: <explanation>
						if (Object.hasOwn(body as any, "message")) {
							// biome-ignore lint/suspicious/noExplicitAny: <explanation>
							errorBody = (body as any).message;
						} else {
							errorBody = JSON.stringify(response.body);
						}
					} catch {
						errorBody = "Could not read error response body";
					}

					// For other status codes, throw a generic error that will trigger retry
					throw new Error(
						`HTTP Error: ${statusCode} ${response.statusMessage}`,
					);
				}
			} catch (error) {
				// Log all errors (both network errors and HTTP errors)
				l.error(`Webhook error: ${(error as Error).message}`, {
					error,
				});

				// Re-throw to trigger retry
				throw error;
			}
		},
		{
			connection: redis,
			concurrency: 500, // Process 500 webhooks simultaneously
		},
	);

	// Graceful shutdown
	async function shutdown() {
		await worker.close();
		process.exit(0);
	}

	process.on("SIGTERM", shutdown);
	process.on("SIGINT", shutdown);
})().catch((e) => {
	logger.error("Error", e);
});
