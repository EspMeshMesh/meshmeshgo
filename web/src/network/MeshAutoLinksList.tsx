import { DataTable, List } from "react-admin";
import { formatNodeId } from "../utils";

export const MeshAutoLinksList = () => {
        return <List>
            <DataTable bulkActionButtons={false}>
                <DataTable.Col source="from" render={record => formatNodeId(record.from)} />
                <DataTable.Col source="to" render={record => formatNodeId(record.to)} />
                <DataTable.Col source="description" />
                <DataTable.NumberCol label="Cost" source="weight" options={{ style: "percent" }} />
            </DataTable>
        </List>;
    };
