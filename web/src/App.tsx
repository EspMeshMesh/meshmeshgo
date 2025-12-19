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
import { EspHomeServerList } from "./esphome/EspHomeServerList";
import { EsphomeClientsList } from "./esphome/EsphomeClientsList";
import { MeshAutoNodesList } from "./network/MeshAutoNodesList";
import { MeshAutoLinksList } from "./network/MeshAutoLinksList";
import { MeshAutoNodeEdit } from "./network/MeshAutoNodeEdit";
import { ShowGraph } from './showgraph/ShowGraph';

export const App = () => (
    <Admin layout={Layout} dataProvider={dataProvider} title="Mesh Network">
        <Resource name="nodes" list={MeshNodesList} edit={MeshNodeEdit} create={MeshNodeCreate} icon={HubIcon} options={{ label: "Static nodes" }} />
        <Resource name="links" list={MeshLinksList} edit={MeshLinkEdit} create={MeshLinkCreate} icon={LinkIcon} options={{ label: "Static links" }} />
        <Resource name="autoNodes" list={MeshAutoNodesList} edit={MeshAutoNodeEdit} icon={HubIcon} options={{ label: "Dynamic nodes" }}/>
        <Resource name="autoLinks" list={MeshAutoLinksList} icon={LinkIcon} options={{ label: "Dynamic links" }}/>
        <Resource name="esphomeServers" list={EspHomeServerList} options={{ label: "EspHome Servers" }} />
        <Resource name="esphomeConnections" list={EsphomeClientsList} options={{ label: "EspHome Clients" }} />
        <Resource name="neighbors" />
        <CustomRoutes>
            <Route path="/discoverylive" element={<Discovery />} />
            <Route path="/showgraph" element={<ShowGraph />} />
        </CustomRoutes>
    </Admin>
);