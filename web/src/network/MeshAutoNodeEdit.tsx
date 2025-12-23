import { Edit, TextInput, BooleanInput, TabbedForm, NumberInput, FormDataConsumer, Toolbar, SaveButton, DateTimeInput, DeleteButton, Button, useRecordContext, useNotify } from "react-admin";
import SettingsIcon from '@mui/icons-material/Settings';
import EditNoteIcon from '@mui/icons-material/EditNote';
import ReplayIcon from '@mui/icons-material/Replay';
import DeleteIcon from '@mui/icons-material/Delete';
import { Typography } from "@mui/material";
import { formatNodeId } from "../utils";

const RebootButton  = () => {
    const record = useRecordContext();
    const notify = useNotify();

    const handleClick = () => {
        fetch(`/api/v1/nodeCommands/${record?.id}/reboot`, {
            method: 'GET',
        }).then((r: Response) => {
            if (r.ok) {
                notify("Reboot successful");
            } else {
                return r.json().then((error: { message: string }) => {
                    throw new Error(error.message);
                });
            }
        }).catch((e) => {
            notify("Reboot failed: " + e.message, { type: "error" });
        });
    }

    return <Button label="Reboot" color="primary" variant="contained" onClick={handleClick}>
        <ReplayIcon />
    </Button>;
}

const CreateToolbar = () => {
    return (
        <Toolbar>
            <SaveButton label="Save changes" color="primary" variant="contained" icon={<EditNoteIcon />} />
            <Typography variant="h6" sx={{ flexGrow: 1 }}></Typography>
            <RebootButton />
            <Typography variant="h6" sx={{ width: 16 }}></Typography>
            <DeleteButton label="Delete" color="error" variant="contained" icon={<DeleteIcon />} />
        </Toolbar>
    );
}

export const MeshAutoNodeEdit = () => {
    return (
        <Edit mutationMode="pessimistic">
            <TabbedForm toolbar={<CreateToolbar />}>
                <TabbedForm.Tab label="Local graph information" icon={<EditNoteIcon />} iconPosition="start" sx={{ maxWidth: '40em', minHeight: 48 }}>
                    <TextInput source="id" format={v => formatNodeId(v)} disabled />
                    <TextInput source="tag" label="Host name" />
                    <TextInput source="firmrev" label="Firmware" disabled />
                    <TextInput source="libvers" label="EspMeshMesh version" disabled />
                    <DateTimeInput source="comptime" label="Compile time" disabled />
                    <DateTimeInput source="last_seen" label="Last seen" parse={(value: string) => (value ? new Date(value) : value === '' ? null : value)} disabled/>
                    <BooleanInput source="in_use" />
                </TabbedForm.Tab>
                <TabbedForm.Tab label="Remote information" icon={<SettingsIcon />} iconPosition="start" sx={{ maxWidth: '40em', minHeight: 48 }}>
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
            </TabbedForm>
        </Edit>
    );
};
