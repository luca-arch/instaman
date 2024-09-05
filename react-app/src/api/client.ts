import type {
  Account,
  CopyJob,
  CopyJobMetadata,
  Job,
  JobType,
  User,
} from "src/api/types";
import {
  transformAccount,
  transformCopyJob,
  transformJob,
  transformUser,
} from "./transformer";

// Base URL of go-instaman APIs.
const baseURL = new URL("/instaman", location.origin);

// Give go-instaman up to 90 seconds to respond, as instaproxy may take some time to establish a session.
const maxRequestTime = 90;

// Default HTTP headers to send with each request.
const headers = {
  Accept: "application/json",
  "Content-Type": "application/json",
};

type CreateCopyJobParams = {
  label: string;
  metadata: CopyJobMetadata;
  nextRun?: Date;
  type: JobType.CopyFollowers | JobType.CopyFollowing;
};

type FindCopyJobParams = {
  direction: "followers" | "following";
  userID: number;
  withPage?: number;
};

// type FindJobParams = {
//   checksum?: string;
//   id?: number;
//   jobType?: string;
//   state?: string;
// };

type FindJobsParams = {
  order?:
    | "-last_run"
    | "last_run"
    | "-next_run"
    | "next_run"
    | "-state"
    | "state"
    | "-label"
    | "label";
  page?: number;
  state?: string;
  type?: string;
};

// Build a GET query params string, including the "?" if non-empty.
const queryArgsBuilder = (params: Record<string, unknown>): string => {
  if (!params) {
    return "";
  }

  const args = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    switch (true) {
      case value === undefined:
        break;
      case value === null:
        args.append(key, "null");
        break;
      case typeof value === "string":
        args.append(key, value);
        break;
      case typeof value === "number":
        args.append(key, `${value}`);
        break;
      case typeof value === "boolean":
        args.append(key, value ? "true" : "false");
        break;
      case value instanceof Date:
        args.append(key, value.toISOString());
        break;
      default:
        throw new Error(`Invalid type for ${key}`);
    }
  }

  return args.size ? "?" + args.toString() : "";
};

// Perform a GET request.
async function httpGet<T>(
  url: string,
  params: Record<string, unknown> = {},
): Promise<T> {
  return await httpSend(url + queryArgsBuilder(params), undefined, "GET");
}

// Perform an HTTP request.
async function httpSend<T>(
  url: string,
  data = {},
  method = "POST",
): Promise<T> {
  const req = {
    headers,
    method,
    signal: AbortSignal.timeout(maxRequestTime * 1000),
  };

  if (method !== "GET" && method !== "HEAD") {
    Object.assign(req, {
      body: data !== undefined ? JSON.stringify(data) : undefined,
    });
  }

  const res = await fetch(baseURL.toString() + url, req);
  if (res.ok) {
    return await res.json();
  }

  console.error(res);

  switch (res.status) {
    case 404:
      throw new Error(`the resource ${url} does not exist`);
    case 500:
      throw new Error(`go-instaman internal server error`);
    case 502:
      throw new Error(`instaproxy internal server error`);
    default:
      throw new Error(res.statusText);
  }
}

// Create a new CopyJob, and then return it.
export const createCopyJob = async (
  params: CreateCopyJobParams,
): Promise<CopyJob> =>
  await httpSend<CopyJob>("/jobs/copy", params).then((job) => {
    return transformCopyJob(job, false);
  });

// Find a CopyJob that was created for the given IG user.
export const findCopyJob = async (
  params: FindCopyJobParams,
): Promise<CopyJob | null> =>
  await httpGet<CopyJob>("/jobs/copy", params).then((job) => {
    if (!job) {
      return null;
    }

    return transformCopyJob(job, typeof params.withPage !== "number");
  });

// Retrieve a paginated list of existing Jobs.
export const findJobs = async (params?: FindJobsParams): Promise<Job[]> =>
  await httpGet<Job[]>("/jobs/all", params).then((jobs) =>
    jobs.map(transformJob),
  );

// Get the main account information.
export const getAccount = async (): Promise<Account> =>
  await httpGet<Account>("/instagram/me").then(transformAccount);

// Find an IG user by ID.
export const getUserByID = async (id: number): Promise<User> =>
  await httpGet<User>(`/instagram/account-id/${id}`).then(transformUser);

// Find an IG user by name.
export const getUserByName = async (name: string): Promise<User> =>
  await httpGet<User>(`/instagram/account/${name}`).then(transformUser);
