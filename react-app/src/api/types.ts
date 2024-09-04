// These type definitions are strictly dependent on go-instaman structs.

/**
 * The account that is logged in on instaproxy app.
 */
export type Account = {
  biography: string;
  fullName: string;
  handler: string;
  id: number;
  pictureURL?: URL;
};

/**
 * A job of type "copy-followers" or "copy-following", with `results` and `resultCount`.
 */
export type CopyJob = {
  checksum: string;
  id: number;
  label: string;
  lastRun: Date | null;
  metadata: CopyJobMetadata;
  nextRun: Date | null;
  resultsCount: number;
  results?: User[];
  state: JobStatus;
  type: JobType.CopyFollowers | JobType.CopyFollowing;
};

/**
 * CopyJob metadata.
 */
export type CopyJobMetadata = {
  frequency: "daily" | "weekly";
  userID: number;
};

/**
 * Allowed job stati.
 */
export enum JobStatus {
  New = "new",
  Paused = "paused",
}

/**
 * Allowed job types.
 */
export enum JobType {
  CopyFollowers = "copy-followers",
  CopyFollowing = "copy-following",
}

/**
 * A generic job descriptor.
 */
export type Job = {
  checksum: string;
  id: number;
  label: string;
  lastRun: Date | null;
  metadata: Record<string, unknown>;
  nextRun: Date | null;
  state: JobStatus;
  type: JobType;
};

/**
 * An Instagram user.
 */
export type User = {
  fullName: string;
  handler: string;
  id: number;
  pictureURL?: URL;
};
