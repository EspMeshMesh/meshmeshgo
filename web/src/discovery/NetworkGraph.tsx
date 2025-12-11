import { useEffect, useRef, useState } from 'react';
import { useGetList } from 'react-admin';
import ForceGraph2D from 'react-force-graph-2d';

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
  id: string;
  from: number;
  to: number;
  weight: number;
};

export const NetworkGraph = () => {
  const fgRef = useRef();
  const [data, setData] = useState<GraphData>({ nodes: [{ id: 0, label: '', is_local: true }], links: [] });
  
  const { data: networkNodes } = useGetList<NetworkNode>('nodes', { });
  const { data: networkLinks, total: _1, isPending: _2, error: _3, refetch: refetchLinks } = useGetList<NetworkLink>('links', { }, { enabled: false });

  useEffect(() => {
    const fg = fgRef.current;
    if(fg) fg.d3Force('link').distance((link: GraphDataLink) => 10 + link.value * 100);
  }, []);

  useEffect(() => {
    if (networkNodes) {
      const nodes = networkNodes.map((node: any) => ({
        id: node.id.toString(),
        label: node.tag,
        is_local: node.is_local
      }));
      console.log('useEffect.nodes', nodes);
      setData({ nodes: nodes, links: [] });
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
      width={window.innerWidth-300}
      height={450}
      backgroundColor="black"
      graphData={data}
      nodeLabel={node => node.label}
      nodeColor={node => node.is_local ? 'yellow' : 'blue'}
      linkColor={link => 'gray'}
      linkLabel={link => (link.value * 100).toString()+'%'}
    />
  );
};