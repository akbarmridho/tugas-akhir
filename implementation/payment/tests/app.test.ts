import dayjs from "dayjs";
import { jest } from "@jest/globals";
import { app } from "../src/http/app.js";
import { redis } from "../src/common/redis.js";
import { Cluster } from "ioredis";

const mockQueueAdd = jest
	.fn<
		(
			name: string,
			data: { id: string },
			opts: { jobId: string; delay: number },
		) => Promise<{ id: string }>
	>()
	.mockResolvedValue({ id: "mock-job-id" });

const mockIsDelayed = jest.fn<() => Promise<boolean>>().mockResolvedValue(true);
const mockPromote = jest.fn<() => Promise<void>>().mockResolvedValue();

const mockGetJob = jest
	.fn<
		() => Promise<{
			isDelayed: () => Promise<boolean>;
			promote: () => Promise<void>;
		}>
	>()
	.mockResolvedValue({
		isDelayed: mockIsDelayed,
		promote: mockPromote,
	});

const mockExportMetrics = jest
	.fn<() => Promise<string>>()
	.mockResolvedValue("mock metrics data");

// Mock the queue
jest.mock("@/src/http/queue.js", () => ({
	queue: {
		add: mockQueueAdd,
		getJob: mockGetJob,
		exportPrometheusMetrics: mockExportMetrics,
	},
}));

// Mock logger
jest.mock("@/src/common/logger.js", () => ({
	logger: {
		info: jest.fn(),
		error: jest.fn(),
	},
}));

// Mock utils
jest.mock("@/src/common/utils.js", () => ({
	generateId: jest.fn<() => string>().mockReturnValue("mock-invoice-id"),
	instanceId: "mock-instance-id",
}));

// Explicitly type the mocked Redis instance
const mockRedis = redis as jest.Mocked<Cluster>;

describe("Payment Service API", () => {
	beforeEach(() => {
		// Clear all mocks before each test
		jest.clearAllMocks();

		// Clear redis mock data
		mockRedis.flushall();
	});

	describe("GET /invoices/:id", () => {
		it("should return 404 when invoice not found", async () => {
			// Don't set any data in redis-mock, so get will return null

			const res = await app.request("/invoices/non-existent-id");
			expect(res.status).toBe(404);

			const body = await res.json();
			expect(body).toEqual({ message: "Invoice not found" });
		});

		it("should return invoice when found", async () => {
			// Create mock invoice data
			const mockInvoice = {
				id: "invoice-123",
				amount: 100,
				description: "Test invoice",
				externalId: "ext-123",
				createdAt: new Date().toISOString(),
				expiredAt: dayjs().add(10, "minute").toDate().toISOString(),
				paidAt: null,
				paidAmount: null,
				status: "pending",
			};

			// Setup Redis mock to return the invoice
			await new Promise((resolve) => {
				mockRedis.set(
					"invoices:invoice-123",
					JSON.stringify(mockInvoice),
					resolve,
				);
			});

			const res = await app.request("/invoices/invoice-123");
			expect(res.status).toBe(200);

			const body = await res.json();
			expect(body).toEqual(mockInvoice);
		});
	});

	describe("POST /invoices", () => {
		it("should create a new invoice", async () => {
			const payload = {
				amount: 150,
				description: "New test invoice",
				externalId: "ext-456",
			};

			const res = await app.request("/invoices", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
			});

			expect(res.status).toBe(200);

			const body = await res.json();
			expect(body).toMatchObject({
				id: "mock-invoice-id",
				amount: payload.amount,
				description: payload.description,
				externalId: payload.externalId,
				status: "pending",
			});

			// Verify queue was called correctly
			expect(mockQueueAdd).toHaveBeenCalledWith(
				"webhook",
				{ id: "mock-invoice-id" },
				expect.objectContaining({
					jobId: "mock-invoice-id",
					delay: expect.any(Number),
				}),
			);

			// Verify data was saved to redis
			const savedInvoice = await new Promise((resolve) => {
				mockRedis.get("invoices:mock-invoice-id", (err, data) =>
					resolve(data ? JSON.parse(data) : null),
				);
			});

			expect(savedInvoice).toMatchObject({
				id: "mock-invoice-id",
				status: "pending",
			});
		});
	});

	describe("POST /invoices/:id/payment", () => {
		it("should handle successful payment", async () => {
			// Create mock invoice
			const mockInvoice = {
				id: "invoice-123",
				amount: 100,
				description: "Test invoice",
				externalId: "ext-123",
				createdAt: new Date().toISOString(),
				expiredAt: dayjs().add(10, "minute").toDate().toISOString(),
				paidAt: null,
				paidAmount: null,
				status: "pending",
			};

			// Store mock invoice in redis
			await new Promise((resolve) => {
				mockRedis.set(
					"invoices:invoice-123",
					JSON.stringify(mockInvoice),
					resolve,
				);
			});

			const res = await app.request("/invoices/invoice-123/payment", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ mode: "success" }),
			});

			expect(res.status).toBe(200);

			const body = await res.json();
			expect(body).toMatchObject({
				id: "invoice-123",
				status: "paid",
				paidAmount: 100,
			});
			expect(body.paidAt).toBeTruthy();

			// Verify job was promoted
			expect(mockGetJob).toHaveBeenCalledWith("invoice-123");
			const job = await mockGetJob.mock.results[0].value;
			// biome-ignore lint/suspicious/noExplicitAny: <explanation>
			expect((job as any).promote).toHaveBeenCalled();
		});

		it("should handle failed payment", async () => {
			// Create mock invoice
			const mockInvoice = {
				id: "invoice-123",
				amount: 100,
				description: "Test invoice",
				externalId: "ext-123",
				createdAt: new Date().toISOString(),
				expiredAt: dayjs().add(10, "minute").toDate().toISOString(),
				paidAt: null,
				paidAmount: null,
				status: "pending",
			};

			// Store mock invoice in redis
			await new Promise((resolve) => {
				mockRedis.set(
					"invoices:invoice-123",
					JSON.stringify(mockInvoice),
					resolve,
				);
			});

			const res = await app.request("/invoices/invoice-123/payment", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ mode: "failure" }),
			});

			expect(res.status).toBe(200);

			const body = await res.json();
			expect(body).toMatchObject({
				id: "invoice-123",
				status: "failed",
				paidAmount: null,
				paidAt: null,
			});
		});

		it("should return 404 when invoice not found", async () => {
			const res = await app.request("/invoices/non-existent-id/payment", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ mode: "success" }),
			});

			expect(res.status).toBe(404);
			expect(await res.json()).toEqual({ message: "Invoice not found" });
		});

		it("should reject payment for expired invoices", async () => {
			// Create mock expired invoice
			const mockInvoice = {
				id: "invoice-123",
				amount: 100,
				description: "Test invoice",
				externalId: "ext-123",
				createdAt: new Date().toISOString(),
				expiredAt: dayjs().subtract(1, "hour").toDate().toISOString(),
				paidAt: null,
				paidAmount: null,
				status: "pending",
			};

			// Store mock invoice in redis
			await new Promise((resolve) => {
				mockRedis.set(
					"invoices:invoice-123",
					JSON.stringify(mockInvoice),
					resolve,
				);
			});

			const res = await app.request("/invoices/invoice-123/payment", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ mode: "success" }),
			});

			expect(res.status).toBe(400);
			expect(await res.json()).toEqual({
				message: "Invoice status must be pending",
			});
		});
	});

	describe("Health Endpoint", () => {
		it("should return healthy when Redis is up", async () => {
			// Default mock returns 'PONG' for ping

			const res = await app.request("/health");
			expect(res.status).toBe(200);

			const body = await res.json();
			expect(body).toEqual({
				status: "healthy",
				message: "Node healthy",
			});
		});

		it("should return unhealthy when Redis returns unexpected response", async () => {
			// Override ping mock for this test
			mockRedis.ping.mockResolvedValueOnce("UNEXPECTED");

			const res = await app.request("/health");
			expect(res.status).toBe(500);

			const body = await res.json();
			expect(body).toEqual({
				status: "unhealthy",
				message: "Redis cluster returned unexpected response",
			});
		});

		it("should return unhealthy when Redis connection fails", async () => {
			// Override ping mock to simulate failure
			mockRedis.ping.mockRejectedValueOnce(new Error("Connection failed"));

			const res = await app.request("/health");
			expect(res.status).toBe(503);

			const body = await res.json();
			expect(body).toEqual({
				status: "unhealthy",
				message: "Redis cluster connection failed",
			});
		});
	});
});
