import { Create, NumberInput, required, SimpleForm, TextInput } from "react-admin";
import { formatNodeId, parseNodeId } from "../utils";


export const MeshLinkCreate = () => {
    return <Create mutationMode="pessimistic">
        <SimpleForm>
            <TextInput source="from" format={v => formatNodeId(v)} validate={required()} parse={v => parseNodeId(v)} />
            <TextInput source="to" format={v => formatNodeId(v)} validate={required()} parse={v => parseNodeId(v)} />
            <NumberInput source="weight" validate={required()} min={0} max={100} step={5} format={v => v * 100} parse={v => v / 100} />
        </SimpleForm>
    </Create>;
};