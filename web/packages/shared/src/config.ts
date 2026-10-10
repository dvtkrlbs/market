export class ApiConfigurationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ApiConfigurationError";
  }
}

export function getApiBaseUrl(): string {
  const value = import.meta.env.VITE_API_BASE_URL?.trim();
  if (!value) {
    throw new ApiConfigurationError(
      "API URL is not configured. Set VITE_API_BASE_URL in web/.env and restart the development server.",
    );
  }

  let url: URL;
  try {
    url = new URL(value);
  } catch {
    throw new ApiConfigurationError("VITE_API_BASE_URL must be an absolute HTTP(S) URL.");
  }
  if (
    !["http:", "https:"].includes(url.protocol) ||
    url.username ||
    url.password ||
    url.search ||
    url.hash
  ) {
    throw new ApiConfigurationError(
      "VITE_API_BASE_URL must be an HTTP(S) URL without credentials, query, or fragment.",
    );
  }
  return url.href.replace(/\/+$/, "");
}
