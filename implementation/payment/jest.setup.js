import { jest } from "@jest/globals";

jest.mock("ioredis", () => import("ioredis-mock"));
