import { useRef, useState, useEffect } from "react";
import ForceGraph2D from 'react-force-graph-2d';
import { useGetList } from "react-admin";


type GraphDataNode = {
    id: number;
    label: string;
    is_local: boolean;
};

type GraphDataLink = {
    source: number;
    target: number;
    value: number;
};

type GraphData = {
    nodes: GraphDataNode[];
    links: GraphDataLink[];
};

type NetworkNode = {
    id: number;
    tag: string;
    is_local: boolean;
};

type NetworkLink = {
    id: number;
    from: number;
    to: number;
    weight: number;
};

interface NetworkGraphProps {
    networkType: number;
}
  
export const NetworkGraph = ({networkType}: NetworkGraphProps) => {
    const fgRef = useRef();
    const [data, setData] = useState<GraphData>({ nodes: [{ id: 0, label: '', is_local: true }], links: [] });

    const { data: networkNodes } = useGetList<NetworkNode>(networkType === 1 ? 'autoNodes' : 'nodes', { });
    const { data: networkLinks, total, isPending, error, refetch: refetchLinks } = useGetList<NetworkLink>(networkType === 1 ? 'autoLinks' : 'links', { }, { enabled: false });

    useEffect(() => {
        const fg = fgRef.current;
        if(fg) fg.d3Force('link').distance((link: GraphDataLink) => 10 + link.value * 100);
    }, []);

    useEffect(() => {
        if (networkNodes) {
            refetchLinks();
        }
    }, [networkNodes]);

    useEffect(() => {
        if (networkLinks && networkNodes) {
            setData({ 
                nodes: networkNodes.map((node: any) => ({ id: node.id, label: node.tag, is_local: node.is_local })), 
                links: networkLinks.map((link: any) => ({ source: link.from, target: link.to, value: link.weight })) 
            });
        }
    }, [networkLinks]);

    return (
        <ForceGraph2D 
            ref={fgRef}
            graphData={data}
            nodeLabel={node => node.label}
            nodeColor={node => node.is_local ? 'yellow' : 'blue'}
            linkLabel={link => (link.value * 100).toString()+'%'}
            width={window.innerWidth-300}
            height={650}
        />
    );
};

