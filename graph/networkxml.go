package graph

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"leguru.net/m/v2/graphml"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

func (g *Network) readGraph(filename string) error {
	xmlFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	gml := graphml.NewGraphML("meshmesh network")
	err = gml.Decode(xmlFile)
	if err != nil {
		return err
	}

	for i, gr := range gml.Graphs {
		if i == 0 {
			logger.Log().WithFields(logrus.Fields{"description": gr.Description}).Info("found graph")
			for _, n := range gr.Nodes {
				descr := n.Description
				attrs, err := n.GetAttributes()
				if err != nil {
					return err
				}

				id, err := utils.ParseNodeId(n.ID)
				if err != nil {
					return err
				}

				if len(descr) == 0 {
					descr, _ = attrs["tag"].(string)
				}

				inuse, ok := attrs["inuse"].(bool)
				if !ok {
					inuse, err = strconv.ParseBool(attrs["inuse"].(string))
					if err != nil {
						return err
					}
				}

				firmware := attrs["firmware"].(string)

				dev := NewNodeDevice(id, inuse, descr)
				dev.Device().SetFirmware(firmware)
				g.AddNode(dev)
			}

			for _, e := range gr.Edges {
				attrs, err := e.GetAttributes()
				if err != nil {
					return err
				}

				src, err := utils.ParseNodeId(e.Source)
				if err != nil {
					return err
				}
				dst, err := utils.ParseNodeId(e.Target)
				if err != nil {
					return err
				}

				var weight float64
				_weight32, ok := attrs["weight"].(float32)
				if ok {
					weight = float64(_weight32)
				} else {
					weight, ok = attrs["weight"].(float64)
					if !ok {
						weight, err = strconv.ParseFloat(attrs["weight"].(string), 32)
						if err != nil {
							return err
						}
					}
				}

				g.SetWeightedEdge(g.NewWeightedEdge(g.Node(src), g.Node(dst), weight))
			}
		}
	}
	return nil
}

func (g *Network) writeGraph(filename string) error {
	gml := graphml.NewGraphML("meshmesh network")

	gml.RegisterKey(graphml.KeyForNode, "inuse", "is node in use", reflect.Bool, true)
	gml.RegisterKey(graphml.KeyForNode, "discover", "state variable for discovery", reflect.Bool, false)
	gml.RegisterKey(graphml.KeyForNode, "buggy", "state variable fr functional status", reflect.Bool, false)
	gml.RegisterKey(graphml.KeyForNode, "firmware", "the node firmware revision", reflect.String, "")
	gml.RegisterKey(graphml.KeyForEdge, "weight", "the node firmware revision", reflect.Float32, 0.0)
	gml.RegisterKey(graphml.KeyForEdge, "weight2", "the node firmware revision", reflect.Float32, 0.0)

	gr, err := gml.AddGraph("the graph", graphml.EdgeDirectionDirected, map[string]interface{}{})
	if err != nil {
		return err
	}

	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node().(NodeDevice)

		attributes := map[string]interface{}{
			"inuse":      node.Device().InUse(),
			"discovered": node.Device().Discovered(),
			"firmware":   node.Device().Firmware(),
		}

		gr.AddNode(attributes, utils.FmtNodeId(node.ID()), node.Device().Tag())
	}

	edges := g.WeightedEdges()
	for edges.Next() {
		edge := edges.WeightedEdge()
		from := edge.From().(NodeDevice)
		to := edge.To().(NodeDevice)

		n1 := gr.GetNode(utils.FmtNodeId(from.ID()))
		n2 := gr.GetNode(utils.FmtNodeId(to.ID()))

		attributes := map[string]interface{}{
			"weight": math.Floor(edge.Weight()*100) / 100,
		}

		description := fmt.Sprintf("from %s:[%s] to %s:[%s]", from.Device().Tag(), utils.FmtNodeId(from.ID()), to.Device().Tag(), utils.FmtNodeId(to.ID()))
		gr.AddEdge(n1, n2, attributes, graphml.EdgeDirectionDefault, description)
	}

	xmlFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer xmlFile.Close()
	err = gml.Encode(xmlFile, true)
	return err
}
