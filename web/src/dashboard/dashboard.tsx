import { Card, CardContent, Grid, Typography } from "@mui/material";
import { Title } from "react-admin";
import { StatCard, StatCardProps } from "./statcard";

const data: StatCardProps[] = [
    {
      title: 'Active dynamic nodes',
      value: '0',
      trend: 'up',
    },
    {
      title: 'Active static links',
      value: '0',
      trend: 'down',
    },
    {
      title: 'Active API connections',
      value: '0',
      trend: 'neutral',
    },
  ];

export const Dashboard = () => (
    <Card>
        <Title title="Welcome to the MeshMeshGo Admin" />
        <Typography component="h2" variant="h6" sx={{ mb: 2, padding: "1em 0 0 1em" }}>Overview</Typography>
        <CardContent>
            <Grid container spacing={2} columns={12}>
                {data.map((card, index) => (
          <Grid key={index} size={{ xs: 12, sm: 6, lg: 3 }}>
            <StatCard {...card} />
          </Grid>
        ))}
            </Grid>
        </CardContent>
    </Card>
);