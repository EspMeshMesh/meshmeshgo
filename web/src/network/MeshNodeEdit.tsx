import { Edit, TextInput, BooleanInput, TabbedForm, List, DataTable, useGetRecordId, NumberInput, FormDataConsumer, Toolbar, SaveButton, DeleteButton, FileField, FileInput } from "react-admin";
import SettingsIcon from '@mui/icons-material/Settings';
import DeleteIcon from '@mui/icons-material/Delete';
import EditNoteIcon from '@mui/icons-material/EditNote';
import { Typography } from "@mui/material";
import { formatNodeId } from "../utils";

const CreateToolbar = () => {
    return (
        <Toolbar>
            <SaveButton label="Save changes" color="primary" variant="contained" icon={<EditNoteIcon />} />
            <Typography variant="h6" sx={{ flexGrow: 1 }}></Typography>
            <DeleteButton label="Delete" color="error" variant="contained" icon={<DeleteIcon />} />
        </Toolbar>
    );
}

export const MeshNodeEdit = () => {
    const recordId = useGetRecordId();

    return <Edit mutationMode="pessimistic">
        <TabbedForm toolbar={<CreateToolbar />}>
            <TabbedForm.Tab label="General" icon={<EditNoteIcon />} iconPosition="start" sx={{ maxWidth: '40em', minHeight: 48 }}>
                <TextInput source="id" format={v => formatNodeId(v)} disabled />
                <TextInput source="tag" />
                <BooleanInput source="in_use" />
            </TabbedForm.Tab>
            <TabbedForm.Tab label="Configuration" icon={<SettingsIcon />} iconPosition="start" sx={{ maxWidth: '40em', minHeight: 48 }}>
                <TextInput source="error" format={v => v?.length > 0 ? v : "No error"} readOnly />
                <FormDataConsumer<{ error: string }>>
                    {({ formData }) => (
                        formData.error.length == 0 &&
                        <>
                            <TextInput source="dev_tag" label="Device tag" />
                            <NumberInput source="channel" min={-1} max={11} step={1} label="WIFI channel" />
                            <NumberInput source="tx_power" min={-1} max={20} step={1} label="TX power" />
                            <NumberInput source="groups" min={0} max={255} step={1} label="Groups" />
                        </>
                    )}
                </FormDataConsumer>
                <TextInput source="dev_firmrev" readOnly />
                <TextInput source="binded" format={v => "0x" + v.toString(16).toUpperCase()} readOnly />
                <TextInput source="flags" readOnly />
            </TabbedForm.Tab>
            <TabbedForm.Tab label="Links">
                <List resource="links" actions={false} pagination={false} filter={{ 'any': recordId }}>
                    <DataTable storeKey="links.of.node" bulkActionButtons={false} size="small">
                        <DataTable.Col source="from" render={record => formatNodeId(record.from)} />
                        <DataTable.Col source="to" render={record => formatNodeId(record.to)} />
                        <DataTable.Col source="weight" />
                    </DataTable>
                </List>
            </TabbedForm.Tab>
        </TabbedForm>
    </Edit>;
};
