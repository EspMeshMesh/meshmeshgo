import { BooleanField, DataTable, DateField, EditButton, List } from "react-admin"
import { formatNodeId } from "../utils";


export const MeshAutoNodesList = () => {
    return <List sort={{ field: "id", order: "ASC" }}>
        <DataTable bulkActionButtons={false}>
            <DataTable.Col source="id" render={record => formatNodeId(record.id)} />
            <DataTable.Col source="tag" label="Hostname" />
            <DataTable.Col source="firmrev" label="Firmware" />
            <DataTable.Col source="comptime" label="Compile time">
                <DateField source="comptime" showTime={true} showDate={true} />
            </DataTable.Col>
            <DataTable.Col source="last_seen" label="Last seen">
                <DateField source="last_seen" showTime={true} showDate={true} />
            </DataTable.Col>
            <DataTable.Col source="in_use">
                <BooleanField source="in_use" />
            </DataTable.Col>
            <DataTable.Col source="path" />
            <DataTable.Col>
                <EditButton />
            </DataTable.Col>
        </DataTable>
    </List>;
};