/**
 * Error class for timeout-specific errors
 */
export class TimeoutError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "TimeoutError";
		// Fix prototype chain for older environments
		Object.setPrototypeOf(this, TimeoutError.prototype);
	}
}

/**
 * Waits for a promise to resolve, but will reject if it takes longer than the specified timeout.
 * Similar to Golang's context with deadline.
 *
 * @param promise - The promise to wait for
 * @param timeoutMs - Maximum time to wait in milliseconds
 * @param errorMessage - Custom error message for timeout
 * @returns A promise that resolves with the original promise result or rejects on timeout
 * @throws {TimeoutError} When the operation times out
 */
export function withTimeout<T>(
	promise: Promise<T>,
	timeoutMs: number,
	errorMessage = "Operation timed out",
): Promise<T> {
	// Create a timeout promise that rejects after the specified time
	const timeoutPromise = new Promise<never>((_, reject) => {
		const id = setTimeout(() => {
			clearTimeout(id);
			reject(new TimeoutError(errorMessage));
		}, timeoutMs);
	});

	// Race the original promise against the timeout
	return Promise.race<T>([promise, timeoutPromise]);
}
