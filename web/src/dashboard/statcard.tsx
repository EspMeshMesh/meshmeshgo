import { Card, CardContent, Chip, Stack, Typography } from "@mui/material"

export type StatCardProps = {
    title: string;
    value: string;
    trend: 'up' | 'down' | 'neutral';
};

export const StatCard = ({ title, value, trend }: StatCardProps) => {

const labelColors = {
    up: 'success' as const,
    down: 'error' as const,
    neutral: 'default' as const,
};

const color = labelColors[trend];

return (
<Card variant="outlined" sx={{ height: '100%', flexGrow: 1 }}>
    <CardContent>
        <Typography component="h2" variant="subtitle2" gutterBottom>
            {title}
        </Typography>
        <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="h3" component="div">
                {value}
            </Typography>
            <Typography variant="subtitle1" color="text.secondary">
                <Chip size="small" color={color} label={trend} />
            </Typography>
        </Stack>
    </CardContent>
</Card>
);
};