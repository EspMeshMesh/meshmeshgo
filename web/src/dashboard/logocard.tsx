import { Card, CardContent, Stack, Typography } from "@mui/material";


export const LogoCard = ({ revision, name }: { revision: string, name: string }) => {
    return (
        <Card>
            <CardContent>
                <Typography component="h2" variant="subtitle2" gutterBottom>Application revision {revision}</Typography>
                <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="body1" component="div">Application name: {name}</Typography>
                    <img src="logo.png" alt="Logo" width="64px" height="64px" />
                </Stack>
            </CardContent>
        </Card>
    );
};