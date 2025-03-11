/** @type {import('ts-jest').JestConfigWithTsJest} **/
export default {
  preset: "ts-jest", // Use the ESM preset
  testEnvironment: "node",
  extensionsToTreatAsEsm: [".ts"],
  moduleNameMapper: {
    // Fix relative imports ending in .js
    "^(\\.{1,2}/.*)\\.js$": "$1",
    // Add alias for cleaner imports
    "^@src/(.*)$": "<rootDir>/src/$1",
  },
  transform: {
    "^.+\\.tsx?$": [
      "ts-jest",
      {
        useESM: true, // Explicitly enable ESM
      },
    ],
  },
  transformIgnorePatterns: ["node_modules/(?!(ioredis-mock)/)"], // Ensure ioredis-mock is transformed
  setupFilesAfterEnv: ["<rootDir>/jest.setup.js"],
  rootDir: ".",
  moduleDirectories: ["node_modules", "<rootDir>/src"],
};
