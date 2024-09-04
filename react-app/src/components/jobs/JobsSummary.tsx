import { Alert, Button, Divider, Spin, Table } from "antd";
import { useEffect, useState } from "react";
import type { Job } from "src/api/types";
import { LoadingOutlined } from "@ant-design/icons";
import NewCopyJobModal from "./NewCopyJobModal";
import { findJobs } from "src/api/client";

const { Column } = Table;

interface TableRow extends Job {
  key: React.Key;
}

const renderDateTime = (d: Date | undefined) => {
  if (!d) {
    return <>&ndash;</>;
  }

  return (
    <>
      {d.toLocaleString()} ({Intl.DateTimeFormat().resolvedOptions().timeZone})
    </>
  );
};

const JobsSummary = () => {
  const [copyJobModalOpen, setCopyJobModalOpen] = useState<boolean>(false);
  const [jobs, setJobs] = useState<Job[]>();
  const [errorMsg, setErrorMsg] = useState<string>();

  useEffect(() => {
    if (jobs === undefined) {
      findJobs()
        .then(setJobs)
        .catch((err) => setErrorMsg(`Error while loading jobs! ${err}.`));
    }
  }, [jobs]);

  if (errorMsg) {
    return <Alert message={errorMsg} type="error" showIcon />;
  }

  if (!jobs) {
    return <Spin indicator={<LoadingOutlined spin />} size="large" />;
  }

  const data: TableRow[] = jobs.map((job, key) => ({
    ...job,
    key,
  }));

  return (
    <>
      <Table dataSource={data}>
        <Column title="ID" dataIndex="id" key="id" />
        <Column title="Label" dataIndex="label" key="label" />
        <Column title="Type" dataIndex="type" key="type" />
        <Column title="Status" dataIndex="state" key="state" />
        <Column
          title="Last run"
          dataIndex="lastRun"
          key="lastRun"
          render={renderDateTime}
        />
        <Column
          title="Next run"
          dataIndex="nextRun"
          key="nextRun"
          render={renderDateTime}
        />

        {/* <Column
          title="Tags"
          dataIndex="tags"
          key="tags"
          render={(tags: string[]) => (
            <>
              {tags.map((tag) => {
                let color = tag.length > 5 ? "geekblue" : "green";
                if (tag === "loser") {
                  color = "volcano";
                }
                return (
                  <Tag color={color} key={tag}>
                    {tag.toUpperCase()}
                  </Tag>
                );
              })}
            </>
          )}
        />*/}
        {/* <Column
          title="Action"
          key="action"
          render={(_: unknown, record: TableRow) => (
            <Space size="middle">
              <a>Invite {record.lastName}</a>
              <a>{account.handler}</a>
            </Space>
          )}
        /> */}
      </Table>

      <Divider />

      <Button
        size="large"
        onClick={() => {
          setCopyJobModalOpen(true);
        }}
      >
        New Copy Job
      </Button>

      <NewCopyJobModal
        open={copyJobModalOpen}
        onCancel={() => {
          setCopyJobModalOpen(false);
        }}
        onCreate={() => {
          setCopyJobModalOpen(false);
          setJobs(undefined);
        }}
      />
    </>
  );
};

export default JobsSummary;
