import { Route } from "react-router-dom";
import { Admin, CustomRoutes, Resource } from "react-admin";

import HubIcon from '@mui/icons-material/Hub';
import LinkIcon from '@mui/icons-material/Link';

import { Layout } from "./Layout";
import { MeshNodesList } from "./network/MeshNodesList";
import { dataProvider } from "./dataProvider";
import { MeshLinksList } from "./network/MeshLinksList";
import { MeshLinkEdit } from "./network/MeshLinkEdit";
import { MeshNodeEdit } from "./network/MeshNodeEdit";
import { MeshNodeCreate } from "./network/MeshNodeCreate";
import { MeshLinkCreate } from "./network/MeshLinkCreate";
import { Discovery } from "./discovery/discovery";
import { FirmwareShow } from "./firmware/FirmwareShow";
import { FirmwareEdit } from "./firmware/FirmwareEdit";
import { EspHomeServerList } from "./esphome/EspHomeServerList";
import { EsphomeClientsList } from "./esphome/EsphomeClientsList";

export const App = () => (
    <Admin layout={Layout} dataProvider={dataProvider} title="Mesh Network">
        <Resource name="nodes" list={MeshNodesList} edit={MeshNodeEdit} create={MeshNodeCreate} icon={HubIcon} />
        <Resource name="links" list={MeshLinksList} edit={MeshLinkEdit} create={MeshLinkCreate} icon={LinkIcon} />
        <Resource name="esphomeServers" list={EspHomeServerList} options={{ label: "EspHome Servers" }} />
        <Resource name="esphomeConnections" list={EsphomeClientsList} options={{ label: "EspHome Clients" }} />
        <Resource name="firmware" edit={FirmwareEdit} show={FirmwareShow} icon={HubIcon} />
        <Resource name="neighbors" />
        <CustomRoutes>
            <Route path="/discoverylive" element={<Discovery />} />
        </CustomRoutes>
    </Admin>
);