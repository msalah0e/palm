package graph

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entity represents a node in the knowledge graph.
type Entity struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Observations []string  `json:"observations"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Relation represents a directed edge between two entities.
type Relation struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Graph is the top-level container for entities and relations.
type Graph struct {
	Entities  map[string]*Entity `json:"entities"`
	Relations []*Relation        `json:"relations"`
}

// Stats holds summary counts.
type Stats struct {
	Entities     int
	Relations    int
	Observations int
	Types        int
}

// ShowResult holds the data for displaying an entity with its connections.
type ShowResult struct {
	Entity   *Entity    `json:"entity"`
	Outgoing []ShowEdge `json:"outgoing"`
	Incoming []ShowEdge `json:"incoming"`
}

// ShowEdge represents a connected entity in a show result.
type ShowEdge struct {
	Type   string  `json:"type"`
	Target *Entity `json:"target,omitempty"`
	Source *Entity `json:"source,omitempty"`
}

// SearchResult holds a scored search hit.
type SearchResult struct {
	Entity *Entity `json:"entity"`
	Score  int     `json:"score"`
}

// ─── Encryption (duplicated from internal/vault/file.go) ───

func deriveKey() []byte {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	seed := fmt.Sprintf("palm-graph:%s:%s", hostname, username)
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ─── Storage ───

func graphPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "palm", "graph.enc")
}

// New creates an empty graph.
func New() *Graph {
	return &Graph{
		Entities:  make(map[string]*Entity),
		Relations: make([]*Relation, 0),
	}
}

// Load reads and decrypts the graph from disk. Returns empty graph if file doesn't exist.
func Load() (*Graph, error) {
	data, err := os.ReadFile(graphPath())
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}

	key := deriveKey()
	plaintext, err := decrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("graph decrypt: %w", err)
	}

	g := New()
	if err := json.Unmarshal(plaintext, g); err != nil {
		return nil, fmt.Errorf("graph parse: %w", err)
	}
	if g.Entities == nil {
		g.Entities = make(map[string]*Entity)
	}
	if g.Relations == nil {
		g.Relations = make([]*Relation, 0)
	}
	return g, nil
}

// Save encrypts and writes the graph to disk.
func Save(g *Graph) error {
	plaintext, err := json.Marshal(g)
	if err != nil {
		return err
	}

	key := deriveKey()
	ciphertext, err := encrypt(key, plaintext)
	if err != nil {
		return err
	}

	path := graphPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, ciphertext, 0o600)
}

// ─── CRUD ───

func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// AddEntity creates a new entity. Returns error if it already exists.
func (g *Graph) AddEntity(name, entityType string) error {
	key := normalize(name)
	if key == "" {
		return fmt.Errorf("entity name cannot be empty")
	}
	if _, exists := g.Entities[key]; exists {
		return fmt.Errorf("entity already exists: %s", name)
	}
	now := time.Now()
	g.Entities[key] = &Entity{
		Name:         name,
		Type:         entityType,
		Observations: make([]string, 0),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return nil
}

// GetEntity returns an entity by name (case-insensitive).
func (g *Graph) GetEntity(name string) (*Entity, error) {
	e, ok := g.Entities[normalize(name)]
	if !ok {
		return nil, fmt.Errorf("entity not found: %s", name)
	}
	return e, nil
}

// RemoveEntity deletes an entity and all its relations.
func (g *Graph) RemoveEntity(name string) error {
	key := normalize(name)
	if _, ok := g.Entities[key]; !ok {
		return fmt.Errorf("entity not found: %s", name)
	}
	delete(g.Entities, key)

	// Cascade: remove all relations involving this entity
	filtered := make([]*Relation, 0, len(g.Relations))
	for _, r := range g.Relations {
		if normalize(r.From) != key && normalize(r.To) != key {
			filtered = append(filtered, r)
		}
	}
	g.Relations = filtered
	return nil
}

// AddObservation appends an observation to an entity.
func (g *Graph) AddObservation(name, observation string) error {
	e, err := g.GetEntity(name)
	if err != nil {
		return err
	}
	e.Observations = append(e.Observations, observation)
	e.UpdatedAt = time.Now()
	return nil
}

// RemoveObservation removes an observation by index.
func (g *Graph) RemoveObservation(name string, index int) error {
	e, err := g.GetEntity(name)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(e.Observations) {
		return fmt.Errorf("observation index out of range: %d (have %d)", index, len(e.Observations))
	}
	e.Observations = append(e.Observations[:index], e.Observations[index+1:]...)
	e.UpdatedAt = time.Now()
	return nil
}

// AddRelation creates a directed relation. Both entities must exist.
func (g *Graph) AddRelation(from, relType, to string) error {
	fromKey := normalize(from)
	toKey := normalize(to)

	if _, ok := g.Entities[fromKey]; !ok {
		return fmt.Errorf("entity not found: %s", from)
	}
	if _, ok := g.Entities[toKey]; !ok {
		return fmt.Errorf("entity not found: %s", to)
	}

	// Deduplicate
	for _, r := range g.Relations {
		if normalize(r.From) == fromKey && r.Type == relType && normalize(r.To) == toKey {
			return fmt.Errorf("relation already exists: %s --%s--> %s", from, relType, to)
		}
	}

	g.Relations = append(g.Relations, &Relation{
		From: g.Entities[fromKey].Name,
		To:   g.Entities[toKey].Name,
		Type: relType,
	})
	return nil
}

// RemoveRelation removes a specific relation.
func (g *Graph) RemoveRelation(from, relType, to string) error {
	fromKey := normalize(from)
	toKey := normalize(to)

	for i, r := range g.Relations {
		if normalize(r.From) == fromKey && r.Type == relType && normalize(r.To) == toKey {
			g.Relations = append(g.Relations[:i], g.Relations[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("relation not found: %s --%s--> %s", from, relType, to)
}

// ─── Query ───

// RelationsOf returns outgoing and incoming relations for an entity.
func (g *Graph) RelationsOf(name string) ([]*Relation, []*Relation) {
	key := normalize(name)
	var outgoing, incoming []*Relation
	for _, r := range g.Relations {
		if normalize(r.From) == key {
			outgoing = append(outgoing, r)
		}
		if normalize(r.To) == key {
			incoming = append(incoming, r)
		}
	}
	return outgoing, incoming
}

// Search finds entities matching a query string. Scored: name(100) > type(20) > observation(10).
func (g *Graph) Search(query string) []SearchResult {
	q := strings.ToLower(query)
	var results []SearchResult

	for _, e := range g.Entities {
		score := 0
		nameLower := strings.ToLower(e.Name)
		typeLower := strings.ToLower(e.Type)

		if nameLower == q {
			score += 100
		} else if strings.Contains(nameLower, q) {
			score += 50
		}

		if typeLower == q {
			score += 20
		} else if strings.Contains(typeLower, q) {
			score += 15
		}

		for _, obs := range e.Observations {
			if strings.Contains(strings.ToLower(obs), q) {
				score += 10
				break
			}
		}

		if score > 0 {
			results = append(results, SearchResult{Entity: e, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

// GetStats returns summary statistics.
func (g *Graph) GetStats() Stats {
	totalObs := 0
	types := make(map[string]bool)
	for _, e := range g.Entities {
		totalObs += len(e.Observations)
		if e.Type != "" {
			types[e.Type] = true
		}
	}
	return Stats{
		Entities:     len(g.Entities),
		Relations:    len(g.Relations),
		Observations: totalObs,
		Types:        len(types),
	}
}

// EntityNames returns a sorted list of entity display names.
func (g *Graph) EntityNames() []string {
	names := make([]string, 0, len(g.Entities))
	for _, e := range g.Entities {
		names = append(names, e.Name)
	}
	sort.Strings(names)
	return names
}

// ShowEntity builds the data for displaying an entity with its connections.
func (g *Graph) ShowEntity(name string) (*ShowResult, error) {
	e, err := g.GetEntity(name)
	if err != nil {
		return nil, err
	}

	outgoing, incoming := g.RelationsOf(name)
	result := &ShowResult{Entity: e}

	for _, r := range outgoing {
		target, _ := g.GetEntity(r.To)
		result.Outgoing = append(result.Outgoing, ShowEdge{Type: r.Type, Target: target})
	}
	for _, r := range incoming {
		source, _ := g.GetEntity(r.From)
		result.Incoming = append(result.Incoming, ShowEdge{Type: r.Type, Source: source})
	}
	return result, nil
}

// ─── Export / Import ───

// ExportJSON returns the graph as pretty-printed JSON.
func (g *Graph) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}

// ExportDOT returns the graph in Graphviz DOT format.
func (g *Graph) ExportDOT() string {
	var b strings.Builder
	b.WriteString("digraph palm_graph {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box, style=rounded];\n\n")

	// Sort entity keys for deterministic output
	keys := make([]string, 0, len(g.Entities))
	for k := range g.Entities {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		e := g.Entities[k]
		label := e.Name
		if e.Type != "" {
			label += "\\n(" + e.Type + ")"
		}
		b.WriteString(fmt.Sprintf("  %q [label=%q];\n", k, label))
	}

	b.WriteString("\n")
	for _, r := range g.Relations {
		b.WriteString(fmt.Sprintf("  %q -> %q [label=%q];\n", normalize(r.From), normalize(r.To), r.Type))
	}

	b.WriteString("}\n")
	return b.String()
}

// ImportJSON merges entities and relations from JSON data into this graph.
// New entities are added; existing entities get observations appended.
func (g *Graph) ImportJSON(data []byte) (added, merged, relAdded int, err error) {
	var incoming Graph
	if err = json.Unmarshal(data, &incoming); err != nil {
		return 0, 0, 0, fmt.Errorf("import parse: %w", err)
	}

	for key, ie := range incoming.Entities {
		existing, exists := g.Entities[normalize(key)]
		if exists {
			// Merge: append new observations
			obsSet := make(map[string]bool)
			for _, o := range existing.Observations {
				obsSet[o] = true
			}
			for _, o := range ie.Observations {
				if !obsSet[o] {
					existing.Observations = append(existing.Observations, o)
				}
			}
			existing.UpdatedAt = time.Now()
			merged++
		} else {
			// Add new entity
			if ie.Observations == nil {
				ie.Observations = make([]string, 0)
			}
			if ie.CreatedAt.IsZero() {
				ie.CreatedAt = time.Now()
			}
			if ie.UpdatedAt.IsZero() {
				ie.UpdatedAt = time.Now()
			}
			g.Entities[normalize(key)] = ie
			added++
		}
	}

	// Import relations (deduplicate)
	for _, ir := range incoming.Relations {
		fromKey := normalize(ir.From)
		toKey := normalize(ir.To)

		// Skip if either entity doesn't exist in final graph
		if _, ok := g.Entities[fromKey]; !ok {
			continue
		}
		if _, ok := g.Entities[toKey]; !ok {
			continue
		}

		dup := false
		for _, r := range g.Relations {
			if normalize(r.From) == fromKey && r.Type == ir.Type && normalize(r.To) == toKey {
				dup = true
				break
			}
		}
		if !dup {
			g.Relations = append(g.Relations, ir)
			relAdded++
		}
	}

	return added, merged, relAdded, nil
}

// ─── Visualization ───

// RenderShow produces a terminal tree view of an entity and its connections.
func RenderShow(g *Graph, name string, brandFn, subtleFn, infoFn func(string) string) (string, error) {
	result, err := g.ShowEntity(name)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	// Incoming relations (above the entity)
	for i, edge := range result.Incoming {
		prefix := "  \u251c\u2500\u2500 "
		if i == len(result.Incoming)-1 && len(result.Outgoing) == 0 {
			prefix = "  \u2514\u2500\u2500 "
		}
		sourceName := ""
		sourceType := ""
		if edge.Source != nil {
			sourceName = edge.Source.Name
			sourceType = edge.Source.Type
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s\n", prefix, subtleFn(edge.Type), subtleFn("\u2500\u2500"), brandFn(sourceName)))
		if sourceType != "" {
			b.WriteString(fmt.Sprintf("  \u2502           %s\n", subtleFn(sourceType)))
		}
		b.WriteString("  \u2502\n")
	}

	// Entity center
	b.WriteString(fmt.Sprintf("  \u25cf %s\n", brandFn(result.Entity.Name)))
	if result.Entity.Type != "" {
		b.WriteString(fmt.Sprintf("  \u2502  %s\n", subtleFn(result.Entity.Type)))
	}
	for _, obs := range result.Entity.Observations {
		b.WriteString(fmt.Sprintf("  \u2502  %s\n", infoFn("\""+obs+"\"")))
	}

	// Outgoing relations (below the entity)
	if len(result.Outgoing) > 0 {
		b.WriteString("  \u2502\n")
	}
	for i, edge := range result.Outgoing {
		prefix := "  \u251c\u2500\u2500 "
		if i == len(result.Outgoing)-1 {
			prefix = "  \u2514\u2500\u2500 "
		}
		targetName := ""
		targetType := ""
		if edge.Target != nil {
			targetName = edge.Target.Name
			targetType = edge.Target.Type
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s\n", prefix, subtleFn(edge.Type), subtleFn("\u2500\u2500"), brandFn(targetName)))
		if targetType != "" {
			b.WriteString(fmt.Sprintf("              %s\n", subtleFn(targetType)))
		}
		// Show first observation of target if present
		if edge.Target != nil && len(edge.Target.Observations) > 0 {
			b.WriteString(fmt.Sprintf("              %s\n", infoFn("\""+edge.Target.Observations[0]+"\"")))
		}
	}

	return b.String(), nil
}

// ─── HTML Visualization (Obsidian-like graph view) ───

// ExportHTML returns a self-contained HTML file with a force-directed graph visualization.
// All data is embedded as JSON constants — no external dependencies.
func (g *Graph) ExportHTML() string {
	// Build nodes and edges arrays as JSON for the JS
	type jsNode struct {
		ID   string   `json:"id"`
		Name string   `json:"name"`
		Type string   `json:"type"`
		Obs  []string `json:"obs"`
	}
	type jsEdge struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Type   string `json:"type"`
	}

	nodes := make([]jsNode, 0, len(g.Entities))
	keys := make([]string, 0, len(g.Entities))
	for k := range g.Entities {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		e := g.Entities[k]
		nodes = append(nodes, jsNode{ID: k, Name: e.Name, Type: e.Type, Obs: e.Observations})
	}

	edges := make([]jsEdge, 0, len(g.Relations))
	for _, r := range g.Relations {
		edges = append(edges, jsEdge{Source: normalize(r.From), Target: normalize(r.To), Type: r.Type})
	}

	nodesJSON, _ := json.Marshal(nodes)
	edgesJSON, _ := json.Marshal(edges)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>palm graph</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0a0e17;color:#e0e0e0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;overflow:hidden}
canvas{display:block}
#info{position:fixed;top:16px;left:16px;z-index:10;background:rgba(10,14,23,0.9);border:1px solid rgba(45,182,130,0.3);border-radius:12px;padding:16px 20px;backdrop-filter:blur(10px);font-size:13px;min-width:200px}
#info h2{color:#2DB682;font-size:16px;margin-bottom:8px}
.stat{color:#888;margin:2px 0}
.stat b{color:#ccc}
#tooltip{position:fixed;z-index:20;pointer-events:none;display:none;background:rgba(10,14,23,0.95);border:1px solid rgba(45,182,130,0.5);border-radius:10px;padding:12px 16px;backdrop-filter:blur(10px);font-size:12px;max-width:300px}
.tt-name{color:#2DB682;font-weight:700;font-size:14px}
.tt-type{color:#888;font-style:italic;margin-bottom:4px}
.tt-obs{color:#aaa;margin:2px 0}
#search-box{position:fixed;top:16px;right:16px;z-index:10;background:rgba(10,14,23,0.9);border:1px solid rgba(45,182,130,0.3);border-radius:8px;padding:8px 14px;color:#e0e0e0;font-size:13px;outline:none;width:200px;font-family:inherit}
#search-box::placeholder{color:#555}
#search-box:focus{border-color:#2DB682}
#legend{position:fixed;bottom:16px;left:16px;z-index:10;background:rgba(10,14,23,0.9);border:1px solid rgba(255,255,255,0.06);border-radius:10px;padding:12px 16px;font-size:11px;color:#666}
.leg-row{margin:3px 0;display:flex;align-items:center;gap:8px}
.dot{width:10px;height:10px;border-radius:50%%;display:inline-block}
</style>
</head>
<body>
<div id="info">
  <h2 id="title"></h2>
  <div class="stat"><b id="n-nodes">0</b> entities</div>
  <div class="stat"><b id="n-edges">0</b> relations</div>
  <div class="stat" id="hint" style="margin-top:8px;color:#555;font-size:11px;"></div>
</div>
<input id="search-box" type="text" placeholder="Search entities...">
<div id="tooltip"></div>
<div id="legend"></div>
<canvas id="canvas"></canvas>
<script>
"use strict";
const NODES=%s;
const EDGES=%s;

document.getElementById('title').textContent='palm graph';
document.getElementById('n-nodes').textContent=NODES.length;
document.getElementById('n-edges').textContent=EDGES.length;
document.getElementById('hint').textContent='drag nodes / scroll to zoom / search to filter';

const PALETTE=['#2DB682','#0171E3','#E07C3A','#9B59B6','#E74C3C','#1ABC9C','#F1C40F','#3498DB','#E91E63','#00BCD4'];
const TYPE_COLORS={};
const types=[...new Set(NODES.map(n=>n.type||'default'))].sort();
types.forEach((t,i)=>{TYPE_COLORS[t]=PALETTE[i%%PALETTE.length]});

const legend=document.getElementById('legend');
types.forEach(t=>{
  const row=document.createElement('div');
  row.className='leg-row';
  const dot=document.createElement('span');
  dot.className='dot';
  dot.style.background=TYPE_COLORS[t];
  row.appendChild(dot);
  const lbl=document.createTextNode(' '+(t||'default'));
  row.appendChild(lbl);
  legend.appendChild(row);
});

const canvas=document.getElementById('canvas');
const ctx=canvas.getContext('2d');
let W,H;
function resize(){W=canvas.width=window.innerWidth;H=canvas.height=window.innerHeight}
resize();
window.addEventListener('resize',resize);

const sim={
  nodes:NODES.map(n=>({...n,x:W/2+(Math.random()-0.5)*300,y:H/2+(Math.random()-0.5)*300,vx:0,vy:0,r:6+Math.min(n.obs.length,10)*1.5,highlight:false,hidden:false})),
  edges:EDGES.map(e=>({...e,si:NODES.findIndex(n=>n.id===e.source),ti:NODES.findIndex(n=>n.id===e.target)})).filter(e=>e.si>=0&&e.ti>=0)
};

let camera={x:0,y:0,zoom:1},drag=null,hovered=null;

function tick(){
  const nodes=sim.nodes,edges=sim.edges;
  const k=0.005,repulse=2000,damp=0.85,center=0.001;
  for(const n of nodes){n.vx+=(W/2-n.x)*center;n.vy+=(H/2-n.y)*center}
  for(let i=0;i<nodes.length;i++){
    for(let j=i+1;j<nodes.length;j++){
      let dx=nodes[j].x-nodes[i].x,dy=nodes[j].y-nodes[i].y;
      let d2=dx*dx+dy*dy;if(d2<1)d2=1;
      let f=repulse/d2,fx=dx*f,fy=dy*f;
      nodes[i].vx-=fx;nodes[i].vy-=fy;nodes[j].vx+=fx;nodes[j].vy+=fy;
    }
  }
  const springLen=120;
  for(const e of edges){
    const a=nodes[e.si],b=nodes[e.ti];
    let dx=b.x-a.x,dy=b.y-a.y,d=Math.sqrt(dx*dx+dy*dy)||1;
    let f=(d-springLen)*k,fx=(dx/d)*f,fy=(dy/d)*f;
    a.vx+=fx;a.vy+=fy;b.vx-=fx;b.vy-=fy;
  }
  for(const n of nodes){
    if(n===drag)continue;
    n.vx*=damp;n.vy*=damp;n.x+=n.vx;n.y+=n.vy;
  }
}

function toScreen(x,y){return[(x-camera.x)*camera.zoom+W/2,(y-camera.y)*camera.zoom+H/2]}
function toWorld(sx,sy){return[(sx-W/2)/camera.zoom+camera.x,(sy-H/2)/camera.zoom+camera.y]}

function draw(){
  ctx.clearRect(0,0,W,H);
  for(const e of sim.edges){
    const a=sim.nodes[e.si],b=sim.nodes[e.ti];
    if(a.hidden||b.hidden)continue;
    const[ax,ay]=toScreen(a.x,a.y),[bx,by]=toScreen(b.x,b.y);
    const isHl=hovered&&(a===hovered||b===hovered);
    ctx.beginPath();ctx.moveTo(ax,ay);ctx.lineTo(bx,by);
    ctx.strokeStyle=isHl?'rgba(45,182,130,0.7)':'rgba(255,255,255,0.08)';
    ctx.lineWidth=isHl?2:1;ctx.stroke();
    const angle=Math.atan2(by-ay,bx-ax);
    const tr=b.r*camera.zoom+4;
    const tx=bx-Math.cos(angle)*tr,ty=by-Math.sin(angle)*tr;
    ctx.beginPath();ctx.moveTo(tx,ty);
    ctx.lineTo(tx-8*Math.cos(angle-0.3),ty-8*Math.sin(angle-0.3));
    ctx.lineTo(tx-8*Math.cos(angle+0.3),ty-8*Math.sin(angle+0.3));
    ctx.closePath();ctx.fillStyle=isHl?'rgba(45,182,130,0.7)':'rgba(255,255,255,0.1)';ctx.fill();
    if(isHl){
      const mx=(ax+bx)/2,my=(ay+by)/2;
      ctx.font='10px -apple-system,sans-serif';ctx.fillStyle='#2DB682';ctx.textAlign='center';
      ctx.fillText(e.type,mx,my-6);
    }
  }
  for(const n of sim.nodes){
    if(n.hidden)continue;
    const[sx,sy]=toScreen(n.x,n.y);
    const r=n.r*camera.zoom;
    const col=TYPE_COLORS[n.type||'default']||'#2DB682';
    const isHl=n===hovered||n.highlight;
    if(isHl){
      ctx.beginPath();ctx.arc(sx,sy,r+6,0,Math.PI*2);
      const grad=ctx.createRadialGradient(sx,sy,r,sx,sy,r+6);
      grad.addColorStop(0,col+'44');grad.addColorStop(1,'transparent');
      ctx.fillStyle=grad;ctx.fill();
    }
    ctx.beginPath();ctx.arc(sx,sy,r,0,Math.PI*2);
    ctx.fillStyle=isHl?col:col+'99';ctx.fill();
    ctx.strokeStyle=col;ctx.lineWidth=isHl?2:1;ctx.stroke();
    ctx.font=(isHl?'bold ':'')+Math.max(11,12*camera.zoom)+'px -apple-system,sans-serif';
    ctx.fillStyle=isHl?'#fff':'#bbb';ctx.textAlign='center';
    ctx.fillText(n.name,sx,sy+r+14*camera.zoom);
  }
}

function findNode(sx,sy){
  const[wx,wy]=toWorld(sx,sy);
  for(let i=sim.nodes.length-1;i>=0;i--){
    const n=sim.nodes[i];if(n.hidden)continue;
    const dx=n.x-wx,dy=n.y-wy;
    if(dx*dx+dy*dy<(n.r+4)*(n.r+4))return n;
  }
  return null;
}

canvas.addEventListener('mousedown',e=>{
  const n=findNode(e.clientX,e.clientY);
  if(n){drag=n;drag.vx=0;drag.vy=0}
  else{drag={pan:true,sx:e.clientX,sy:e.clientY,cx:camera.x,cy:camera.y}}
});
canvas.addEventListener('mousemove',e=>{
  if(drag&&drag.pan){
    camera.x=drag.cx-(e.clientX-drag.sx)/camera.zoom;
    camera.y=drag.cy-(e.clientY-drag.sy)/camera.zoom;
  }else if(drag){
    const[wx,wy]=toWorld(e.clientX,e.clientY);drag.x=wx;drag.y=wy;
  }
  const n=findNode(e.clientX,e.clientY);
  hovered=n;
  const tt=document.getElementById('tooltip');
  if(n){
    canvas.style.cursor='pointer';
    // Build tooltip using DOM (safe text content)
    tt.textContent='';
    const nameEl=document.createElement('div');nameEl.className='tt-name';nameEl.textContent=n.name;tt.appendChild(nameEl);
    if(n.type){const typeEl=document.createElement('div');typeEl.className='tt-type';typeEl.textContent=n.type;tt.appendChild(typeEl)}
    if(n.obs&&n.obs.length>0){
      n.obs.forEach(o=>{const obsEl=document.createElement('div');obsEl.className='tt-obs';obsEl.textContent=o;tt.appendChild(obsEl)});
    }
    tt.style.display='block';tt.style.left=(e.clientX+16)+'px';tt.style.top=(e.clientY+16)+'px';
  }else{
    canvas.style.cursor=drag?'grabbing':'default';tt.style.display='none';
  }
});
canvas.addEventListener('mouseup',()=>{drag=null});
canvas.addEventListener('wheel',e=>{
  e.preventDefault();
  const factor=e.deltaY>0?0.9:1.1;
  camera.zoom=Math.max(0.1,Math.min(5,camera.zoom*factor));
},{passive:false});

document.getElementById('search-box').addEventListener('input',function(){
  const q=this.value.toLowerCase();
  for(const n of sim.nodes){
    n.highlight=q&&(n.name.toLowerCase().includes(q)||(n.type||'').toLowerCase().includes(q));
    n.hidden=false;
  }
});

(function loop(){tick();draw();requestAnimationFrame(loop)})();
</script>
</body>
</html>`, string(nodesJSON), string(edgesJSON))
}
