import { Card, CardContent, Stack, Typography } from "@mui/material";


export const LogoCard = () => {
    return (
        <Card>
            <CardContent>
                <Typography component="h2" variant="subtitle2" gutterBottom>Application revision</Typography>
                <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="body1" component="div">MeshMeshGo</Typography>
                    <img src="logo.png" alt="Logo" width="64px" height="64px" />
                </Stack>
            </CardContent>
        </Card>
    );
};