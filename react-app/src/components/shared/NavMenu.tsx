import {
  FieldTimeOutlined,
  PictureOutlined,
  PlusCircleOutlined,
} from "@ant-design/icons";
import { NavState, NavStateSection } from "./nav-state";
import { Menu } from "antd";

type Props = {
  currentSection: NavStateSection;
  setNavState: (s: NavState) => void;
};

// Top navigation menu.
const NavMenu = ({ currentSection, setNavState }: Props) => (
  <>
    <Menu
      selectedKeys={[`${currentSection}`]}
      items={[
        {
          icon: <FieldTimeOutlined />,
          key: NavStateSection.Jobs,
          label: "Jobs",
        },
        {
          icon: <PictureOutlined />,
          key: NavStateSection.Posts,
          label: "Posts",
        },
        {
          icon: <PlusCircleOutlined />,
          key: `more`,
          label: "More",
        },
      ]}
      mode="horizontal"
      onClick={({ key }) => {
        setNavState({
          section: parseInt(key) as NavStateSection,
        });
      }}
      style={{ flex: 1, minWidth: 0 }}
      theme="dark"
    />
  </>
);

export default NavMenu;
