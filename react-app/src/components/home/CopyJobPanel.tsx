import { type Account, type CopyJob, JobType } from "src/api/types";
import { Alert, Button, Col, Row, Skeleton, Space } from "antd";
import { createCopyJob } from "src/api/client";
import { useState } from "react";

type Props = {
  // Main account.
  account: Account;
  // Relationship direction.
  direction: "followers" | "following";
  // Any error occurred while loading the job.
  error?: string;
  // The CopyJob or null (if loading) or false (if nothing to load).
  job: CopyJob | null | false;
  // State setter for children components.
  setJob: (job: CopyJob) => void;
};

/**
 * CopyJobPanel renders the list of account connections (either followers or following).
 * The received job object, should be null if it is still loading by the parent, or false if there was no job to load.
 */
const CopyJobPanel = ({ account, direction, error, job, setJob }: Props) => {
  const [errorMsg, setErrorMsg] = useState<string>();
  const [isLoading, setIsLoading] = useState<boolean>(false);

  if (errorMsg) {
    return <Alert message={errorMsg} type="error" showIcon />;
  }

  const onClick = () => {
    setIsLoading(true);

    createCopyJob({
      label:
        (direction === "followers"
          ? `People following `
          : `People followed by `) + account.fullName,
      metadata: {
        frequency: "daily",
        userID: account.id,
      },
      type:
        direction === "followers"
          ? JobType.CopyFollowers
          : JobType.CopyFollowing,
    })
      .then((job) => setJob(job))
      .catch((err) => setErrorMsg(`Error while creating the job! ${err}.`))
      .finally(() => setIsLoading(false));
  };

  switch (true) {
    case !!error:
      return <Alert message={error} type="error" showIcon />;
    case job === false:
      return (
        <Row>
          <Col xs={24} sm={24} md={12} lg={10} xl={8}>
            <Alert
              action={
                <Space direction="vertical">
                  <Button
                    disabled={isLoading}
                    onClick={onClick}
                    size="small"
                    type="dashed"
                  >
                    Yes
                  </Button>
                </Space>
              }
              description="Would you like to start the sync now?"
              message="Your account is not synchronising!"
              type="info"
            />
          </Col>
        </Row>
      );
    case job === null:
      return <Skeleton active />;
    default:
      return <>{job.resultsCount} connections</>;
  }
};

export default CopyJobPanel;
