import { useRef, useState, useEffect } from "react";
import ForceGraph2D from 'react-force-graph-2d';
import { useGetList } from "react-admin";


type GraphDataNode = {
    id: number;
    label: string;
    color: string;
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
    in_use: boolean;
    is_local: boolean;
    deep_sleep: boolean;
    dev_type: string;
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
                nodes: networkNodes.map((node: any) => ({ id: node.id, label: node.tag, is_local: node.is_local, color: nodeColor(node) })), 
                links: networkLinks.map((link: any) => ({ source: link.from, target: link.to, value: link.weight })) 
            });
        }
    }, [networkLinks]);

    const nodeColor = (node: NetworkNode) => {
        if(node.is_local) return 'gold';
        if (!node.in_use) return 'gray';
        // edge nodes in deep sleep darkcyan
        if (node.deep_sleep && node.dev_type == 'edge') return 'darkcyan';
        // other nodes in deep sleep red as warning
        if (node.deep_sleep) return 'red';
        // edge nodes in use aqua
        if (node.dev_type == 'edge') return 'aqua'
        return 'seagreen';
    }

    return (
        <ForceGraph2D 
            ref={fgRef}
            graphData={data}
            nodeLabel={node => node.label}
            nodeColor={node => node.color}
            linkLabel={link => (link.value * 100).toString()+'%'}
            width={window.innerWidth-300}
            height={650}
        />
    );
};

