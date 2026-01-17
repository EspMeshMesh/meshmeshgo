import { Card, CardContent, Grid, Typography } from "@mui/material";
import { Title, useDataProvider } from "react-admin";
import { StatCard, StatCardProps } from "./statcard";
import { NetworkGraph } from '../showgraph/NetworkGraph';
import { LogoCard } from "./logocard";
import { MyDataProvider } from "../dataProvider";
import { useEffect, useState } from "react";

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

export const Dashboard = () => {
  const dataProvider = useDataProvider<MyDataProvider>();
  const [revision, setRevision] = useState('');
  const [name, setName] = useState('');

  useEffect(() => {
    dataProvider.hello().then((data) => {
      console.log('Dashboard.hello', data);
      setRevision(data.program_revision)
      setName(data.program_name)
    });
  }, []);

  return (
    <Card>
        <Title title="Welcome to the MeshMeshGo Admin" />
        <Typography component="h2" variant="h6" sx={{ mb: 2, padding: "1em 0 0 1em" }}>Overview</Typography>
        <CardContent>
            <Grid container spacing={2} columns={12}>
              <Grid key="logo-card" size={{ xs: 12, sm: 6, lg: 3 }}>
                <LogoCard revision={revision} name={name} />
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
};