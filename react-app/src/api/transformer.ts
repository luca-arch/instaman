import type { Account, CopyJob, Job, User } from "src/api/types";

// Clean and normalise the Account object returned by the API.
export const transformAccount = (account: Account): Account => {
  return {
    ...account,
    // Enforce URL type.
    pictureURL: account.pictureURL ? new URL(account.pictureURL) : undefined,
  };
};

// Clean and normalise the CopyJob object returned by the API.
export const transformCopyJob = (
  job: CopyJob,
  withResults: boolean = true,
): CopyJob => {
  return {
    ...job,
    // Enforce Date types.
    lastRun: job.lastRun ? new Date(job.lastRun) : null,
    nextRun: job.nextRun ? new Date(job.nextRun) : null,
    // Remove null if results were not requested, else enforce an empty list.
    results: withResults ? job.results || [] : undefined,
  };
};

// Clean and normalise the Job object returned by the API.
export const transformJob = (job: Job): Job => {
  return {
    ...job,
    // Enforce Date types.
    lastRun: job.lastRun ? new Date(job.lastRun) : null,
    nextRun: job.nextRun ? new Date(job.nextRun) : null,
  };
};

// Clean and normalise the User object returned by the API.
export const transformUser = (user: User): User => {
  return {
    ...user,
    // Enforce URL type.
    pictureURL: user.pictureURL ? new URL(user.pictureURL) : undefined,
  };
};
