import { FormControl, Grid, InputLabel, MenuItem, Select, Typography } from "@mui/material";
import { Button } from "react-admin";
import { useEffect, useState } from "react";

interface GraphTopBarProps {
    onSelect: (value: number) => void;
}

export const GraphTopBar = ({onSelect}: GraphTopBarProps) => {  
    const [networkType, setNetworkType] = useState(1);

    useEffect(() => {
        console.log('GraphTopBar.useEffect.networkType', networkType);
    }, [networkType]);

    return (
        <Grid container spacing={2}>
            <Grid size={3}>
                <Typography variant="h6">Graph visualization</Typography>
            </Grid>
            <Grid size={3}>
                <FormControl fullWidth>
                    <InputLabel id="network-type-label">Network type</InputLabel>
                    <Select labelId="network-type-label" value={networkType} onChange={(event) => {setNetworkType(event.target.value); onSelect(event.target.value);}}>
                        <MenuItem value={1}>Dynamic network</MenuItem>
                        <MenuItem value={2}>Static network</MenuItem>
                    </Select>
                </FormControl>
            </Grid>
            <Grid size={6}>
            </Grid>
        </Grid>
    );
};