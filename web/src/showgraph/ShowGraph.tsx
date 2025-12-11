import { Card, CardContent } from '@mui/material';
import { Title } from 'react-admin';
import { NetworkGraph } from './NetworkGraph';
import { GraphTopBar } from './GraphTopBar';
import { useState } from 'react';

export const ShowGraph = () => {
    const [networkType, setNetworkType] = useState(1);

    return (
        <Card>
            <Title title="Show Graph" />
            <CardContent>
                <GraphTopBar onSelect={(value) => { console.log('ShowGraph.onSelect', value); setNetworkType(value); }} />
                <NetworkGraph networkType={networkType} />
            </CardContent>
        </Card>
    )
};