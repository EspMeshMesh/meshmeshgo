import { Card, CardContent, Grid, Typography } from "@mui/material";
import { Title } from "react-admin";
import { StatCard, StatCardProps } from "./statcard";
import { NetworkGraph } from '../showgraph/NetworkGraph';
import { LogoCard } from "./logocard";

const data: StatCardProps[] = [
    {
      title: 'Active dynamic nodes',
      value: '1',
      trend: 'up',
    },
    {
      title: 'Active static links',
      value: '1',
      trend: 'up',
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
              <Grid key="logo-card" size={{ xs: 12, sm: 6, lg: 3 }}>
                <LogoCard />
              </Grid>
              {data.map((card, index) => (
                <Grid key={index} size={{ xs: 12, sm: 6, lg: 3 }}>
                  <StatCard {...card} />
                </Grid>
              ))}
                <Grid key="network-graph" size={12}>
                  <NetworkGraph networkType={1} />
                </Grid>
            </Grid>
        </CardContent>
    </Card>
);