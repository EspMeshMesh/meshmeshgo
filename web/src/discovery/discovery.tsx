import { Card, CardContent, Grid, Typography } from "@mui/material";
import { Button, DataTable, List, Title, useCreate, useGetOne } from "react-admin";
import { NetworkGraph } from "./networkgraph";


export const Discovery = () => {
    const { data: discovery } = useGetOne('neighbors/discovery', { id: 0}, { refetchInterval: 2500 });
    
    const StartDiscoveryButton = () => {
        const [create, { isPending }] = useCreate('neighbors/discovery', {data: {}});
        const startDiscoveryHandler = () => {
            create();
        };
        return <Button variant="contained" color="primary" disabled={isPending || discovery?.status === 'running'} label="Start discovery" onClick={startDiscoveryHandler} />
    };

    const RefreshDiscoveryButton = () => {
        const [create, { isPending }] = useCreate('neighbors/discovery', {data: {mode: 'refresh'}});
        const refreshDiscoveryHandler = () => {
            create();
        };
        return <Button variant="contained" color="primary" disabled={isPending || discovery?.status === 'running'} label="Refresh discovery" onClick={refreshDiscoveryHandler} />
    };

    return (
        <Card>
            <Title title="Discovery" />
            <CardContent>
                <Grid container spacing={2}>
                    <Grid size={2}>
                        <StartDiscoveryButton />
                    </Grid>
                    <Grid size={3}>
                        <RefreshDiscoveryButton />
                    </Grid>
                    <Grid size={6}>
                        <Typography variant="h6">Discovery: {discovery?.status}, CurrentId: {"0x" + discovery?.current_id.toString(16).toUpperCase()} Repetitions: {discovery?.repeat}</Typography>
                    </Grid>
                </Grid>
                <Grid>
                    <Grid size={12}>
                        <div style={{ position: 'relative', width: '100%', height: '450px', border: '1px solid #ccc' }}>
                            <NetworkGraph />
                        </div>
                    </Grid>
                </Grid>
                <List resource="neighbors" queryOptions={{ refetchInterval: 2500}}>
                    <DataTable>
                        <DataTable.Col source="id" />
                        <DataTable.Col source="current" />
                        <DataTable.Col source="next" />
                        <DataTable.Col source="delta" />
                    </DataTable>
                </List>
            </CardContent>
        </Card>
    );
};