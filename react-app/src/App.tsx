import { Layout, Spin, Typography, theme } from "antd";
import { NavState, NavStateSection } from "./components/shared/nav-state";
import { useEffect, useState } from "react";
import type { Account } from "src/api/types";
import AccountSummary from "src/components/home/AccountSummary";
import JobsSummary from "./components/jobs/JobsSummary";
import { LoadingOutlined } from "@ant-design/icons";
import NavMenu from "./components/shared/NavMenu";
import UserProfilePicture from "src/components/shared/UserProfilePicture";
import { getAccount } from "src/api/client";

const { Header, Content } = Layout;

const initialNavState: NavState = {
  section: NavStateSection.Home,
};

// Main application panel.
const App: React.FC = () => {
  const [account, setAccount] = useState<Account>();
  const [navState, setNavState] = useState<NavState>(initialNavState);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Load main account information.
  useEffect(() => {
    if (!account) {
      getAccount()
        .then((acc) => setAccount(acc))
        // Best to just show an alert and do nothing else here.
        .catch((err) =>
          window.alert(`Error while loading account information.\n${err}.`),
        );
    }
  }, [account]);

  if (!account) {
    return (
      <>
        <Typography.Title
          style={{ color: colorBgContainer, textAlign: "center" }}
        >
          Fetching account
        </Typography.Title>

        <Spin fullscreen indicator={<LoadingOutlined spin />} size="large" />
      </>
    );
  }

  return (
    <Layout>
      <Header style={{ alignItems: "center", display: "flex" }}>
        <UserProfilePicture
          height="64px"
          onClick={() => {
            setNavState({
              section: NavStateSection.Home,
            });
          }}
          pictureURL={account.pictureURL}
          width="64px"
        />

        <NavMenu currentSection={navState.section} setNavState={setNavState} />
      </Header>

      <Layout>
        <Layout style={{ padding: "0 24px 24px" }}>
          <Content
            style={{
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
              margin: 0,
              minHeight: 280,
              padding: 24,
            }}
          >
            {(() => {
              switch (navState.section) {
                case NavStateSection.Home:
                  return <AccountSummary account={account} />;
                case NavStateSection.Jobs:
                  return <JobsSummary />;
                default:
                  return <>Unknown section {navState.section}</>;
              }
            })()}
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default App;
