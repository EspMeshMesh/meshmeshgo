import type { ReactNode } from "react";
import { Layout as RALayout, CheckForApplicationUpdate, Menu } from "react-admin";
import SearchIcon from '@mui/icons-material/Search';
import SwaggerIcon from '@mui/icons-material/Code';

export const MyMenu = () => (
  <Menu>
      <Menu.ResourceItem name="nodes" />
      <Menu.ResourceItem name="links" />
      <Menu.ResourceItem name="esphomeServers" />
      <Menu.ResourceItem name="esphomeClients" />
      <Menu.Item to="/discoverylive" primaryText="Discovery" leftIcon={<SearchIcon />} />
      <Menu.Item to="" onClick={() => {window.location.href = "/swagger"; }} primaryText="Swagger" leftIcon={<SwaggerIcon />} />
  </Menu>
);

export const Layout = ({ children }: { children: ReactNode }) => (
  <RALayout menu={MyMenu}>
    {children}
    <CheckForApplicationUpdate />
  </RALayout>
);
