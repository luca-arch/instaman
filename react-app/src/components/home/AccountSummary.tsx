import type { Account, CopyJob } from "src/api/types";
import { Divider, Typography } from "antd";
import { useEffect, useState } from "react";
import CopyJobPanel from "./CopyJobPanel";
import { findCopyJob } from "src/api/client";

type Props = {
  account: Account;
};

// Render a panel with main account's information.
const AccountSummary = ({ account }: Props) => {
  const [errorMsg1, setErrorMsg1] = useState<string>();
  const [errorMsg2, setErrorMsg2] = useState<string>();

  // Account followers.
  const [copyFollowers, setCopyFollowers] = useState<CopyJob | null | false>(
    null,
  );
  useEffect(() => {
    if (copyFollowers === null) {
      findCopyJob({
        direction: "followers",
        userID: account.id,
      })
        .then((job) => setCopyFollowers(job || false))
        .catch((err) =>
          setErrorMsg1(`Error while loading your followers list! ${err}.`),
        );
    }
  }, [copyFollowers, account.id]);

  // Account following.
  const [copyFollowing, setCopyFollowing] = useState<CopyJob | null | false>(
    null,
  );
  useEffect(() => {
    if (copyFollowing === null) {
      findCopyJob({
        direction: "following",
        userID: account.id,
      })
        .then((job) => setCopyFollowing(job || false))
        .catch((err) =>
          setErrorMsg2(
            `Error while loading the list of the people you follow! ${err}.`,
          ),
        );
    }
  }, [copyFollowing, account.id]);

  return (
    <>
      <Typography.Title level={2}>
        @{account.handler} ({account.id})
      </Typography.Title>

      <Typography.Title level={3}>
        {account.fullName}
      </Typography.Title>

      {account.biography.split("\n").map((line, key) => (
        <p key={key}>{line}</p>
      ))}

      <Divider>Followers</Divider>

      <CopyJobPanel
        account={account}
        direction="followers"
        error={errorMsg1}
        job={copyFollowers}
        setJob={setCopyFollowers}
      />

      <Divider>Following</Divider>

      <CopyJobPanel
        account={account}
        direction="following"
        error={errorMsg2}
        job={copyFollowing}
        setJob={setCopyFollowing}
      />
    </>
  );
};

export default AccountSummary;
