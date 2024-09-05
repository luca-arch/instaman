import {
  Alert,
  Button,
  Checkbox,
  Form,
  FormProps,
  Input,
  Modal,
  Radio,
  Select,
  Space,
} from "antd";
import { createCopyJob, getUserByName } from "src/api/client";
import { JobType } from "src/api/types";
import { useState } from "react";

type FormData = {
  freq: "daily" | "weekly";
  label: string;
  startNow: boolean;
  type: JobType;
  username: string;
};

type Props = {
  onCancel: () => void;
  onCreate: () => void;
  open: boolean;
};

// Render a form within a modal window that allows the user to create a new CopyJob.
const NewCopyJobModal = ({ onCancel, onCreate, open }: Props) => {
  const [errorMsg, setErrorMsg] = useState<string>();

  const onFinish: FormProps<FormData>["onFinish"] = (form: FormData) => {
    // Find the user ID first.
    getUserByName(form.username)
      .then((user) => {
        // Second request to create the job
        createCopyJob({
          label: form.label,
          metadata: {
            frequency: form.freq,
            userID: user.id,
          },
          nextRun: form.startNow ? new Date() : undefined,
          type: form.type,
        })
          .then(() => onCreate())
          .catch((err) => setErrorMsg(`Error while creating the job! ${err}.`));
      })
      .catch((err) =>
        setErrorMsg(`Error while searching for ${form.username}! ${err}.`),
      );
  };

  return (
    <Modal
      centered
      footer={null}
      onCancel={onCancel}
      open={open}
      title="Create new copy job"
    >
      <Form
        autoComplete="off"
        initialValues={{ remember: true }}
        labelCol={{ span: 6 }}
        name="basic"
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        wrapperCol={{ span: 18 }}
      >
        <Form.Item<FormData>
          label="Username"
          name="username"
          rules={[
            { message: "Please specify a valid user handle!", required: true },
          ]}
        >
          <Input placeholder="Instagram user" />
        </Form.Item>

        <Form.Item<FormData>
          label="Label"
          name="label"
          rules={[{ message: "Please specify a label!", required: true }]}
        >
          <Input placeholder="My copy job" />
        </Form.Item>

        <Form.Item<FormData>
          label="Type"
          name="type"
          rules={[{ message: "Please choose the type!", required: true }]}
        >
          <Radio.Group>
            <Space direction="vertical">
              <Radio value={JobType.CopyFollowers}>Copy followers</Radio>
              <Radio value={JobType.CopyFollowing}>Copy following</Radio>
            </Space>
          </Radio.Group>
        </Form.Item>

        <Form.Item<FormData>
          label="Frequency"
          name="freq"
          rules={[
            {
              message: "Please choose how often the job should run!",
              required: true,
            },
          ]}
        >
          <Select>
            <Select.Option value="daily">Daily</Select.Option>
            <Select.Option value="weekly">Weekly</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item<FormData>
          name="startNow"
          valuePropName="checked"
          wrapperCol={{ offset: 6, span: 18 }}
        >
          <Checkbox>Start immediately</Checkbox>
        </Form.Item>

        {errorMsg && <Alert message={errorMsg} type="error" showIcon />}

        <Form.Item style={{ textAlign: "right" }} wrapperCol={{ span: 24 }}>
          <Button type="primary" htmlType="submit">
            Create
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default NewCopyJobModal;
