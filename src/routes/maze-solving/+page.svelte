<script lang="ts">
	import { onMount } from "svelte";
	import MazeControls from "./MazeControls.svelte";
  import MazeGraph from "./MazeGraph.svelte";
	import MazeGrid from "./MazeGrid.svelte";
	import { Maze } from "./maze";
	import { StackFrontier, type Frontier, type Node, QueueFrontier, PriorityQueueFrontier, GreedyPriorityQueueFrontier } from "./utils";
	import { readable, writable, type Readable, type Writable } from "svelte/store";
	import type { SelectOption } from "@melt-ui/svelte";
	import MultiLangCodeBlock from "$lib/components/MultiLangCodeBlock.svelte";
	import { cpp, javascript, python, rust } from "svelte-highlight/languages";

  const updateMaze = (_: Node) => {
    mazeGridRef.updateGrid();
    mazeGraphRef.updateGraph();
  };

  const createMaze = () => {
    switch ($algorithm.value) {
      case "Depth-first search":
        return new Maze<StackFrontier>(new StackFrontier(), updateMaze);
      case "Breadth-first search":
        return new Maze<QueueFrontier>(new QueueFrontier(), updateMaze);
      case "Greedy best-first search":
        return new Maze<GreedyPriorityQueueFrontier>(new GreedyPriorityQueueFrontier(), updateMaze);
      case "A* search":
        return new Maze<PriorityQueueFrontier>(new PriorityQueueFrontier(), updateMaze);
      default:
        throw new Error("Invalid algorithm");
    }
  };

  const loadMaze = () => {
    const newMaze = createMaze();
    switch ($presetMaze) {
      case 1:
        newMaze.loadFromString(
`A    #     
 # # # # ##
 # #   #   
 # # ### # 
 # #     # 
 # ### ### 
          B`
        );
        break;
      case 2:
        newMaze.loadFromString(
`A          
#### ##### 
     #   # 
 ##### # # 
     # #   
 ### ######
   #      B`
        );
        break;
      case 3:
        newMaze.loadFromString(
`A#         
 # ### ### 
   #   #   
 ### ### # 
   #   # # 
## ### ### 
     #    B`
        );
        break;
      case 4:
        newMaze.loadFromString(
`A   #      
### # #### 
#   #    # 
# # #### # 
# #      # 
# ######## 
#         B`
        );
        break;
      default:
        return;
    }

    maze = readable(newMaze, (set) => {
      set(newMaze);
    });
    if (mazeGridRef) mazeGridRef.updateGrid();
    if (mazeGraphRef) mazeGraphRef.updateGraph();
  };

  let presetMaze: Writable<number> = writable(1);
  let algorithm: Writable<SelectOption<string>> = writable({value: "Depth-first search", label: "Depth-first search"});
  let maze: Readable<Maze<Frontier>> = readable(createMaze(), (set) => {
    set(createMaze());
  });
  let mazeGridRef: MazeGrid;
  let mazeGraphRef: MazeGraph;

  presetMaze.subscribe((value) => {
    loadMaze();
  });

  algorithm.subscribe((value) => {
    maze = readable(createMaze(), (set) => {
      set(createMaze());
    });
    loadMaze();
    if (mazeGridRef) mazeGridRef.updateGrid();
    if (mazeGraphRef) mazeGraphRef.updateGraph();
  });

  loadMaze();
  onMount(() => {
    mazeGridRef.updateGrid();
    mazeGraphRef.updateGraph();
  });
</script>

<div class="container mx-auto">
  <h1 class="mt-16 mb-16 text-6xl font-semibold text-center">
    <span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Maze solving</span>
  </h1>
  <section id="maze">
    <MazeControls maze={maze} presetMaze={presetMaze} algorithm={algorithm} />
    <MazeGrid maze={maze} bind:this={mazeGridRef} />
    <MazeGraph maze={maze} bind:this={mazeGraphRef} />
  </section>
  <section class="my-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">Core concepts of search algorithms</h2>
    <p class="mb-4">
      Searching problems are a very common type of problem in computer science.
      They are used in many applications, such as route planning, scheduling, and games.
    </p>
    <p class="mb-4">
      A search algorithm is an algorithm that explores a graph, starting from a given node,
      to find a node that satisfies a given condition. Nodes define a particular state of the
      problem and the links between nodes are actions that can be performed. The aim of the
      algorithm is to find a sequence of actions that leads to a node that satisfies the condition.
      Of course, this sequence of actions is not necessarily unique and the algorithm must
      hopefully find the best one based on some criteria. For example, the route with the least
      number of actions, or the route with the lowest cost, etc...
    </p>
    <p class="mb-4">
      Here is an implementation of the node class.
    </p>
    <MultiLangCodeBlock code={
      [
        {
          id: 'tab-1',
          title: 'Python',
          lang: python,
          code: `class Node():
  def __init__(self, state, parent, action, g, h):
    self.state = state # for example, a tuple (x, y) representing the position in a maze
    self.parent = parent
    self.action = action
    self.g = g # cost to reach this node
    self.h = h # heuristic value of this node
    self.f = g + h # priority of this node`
        },
        {
          id: 'tab-2',
          title: 'JavaScript',
          lang: javascript,
          code: `class Node {
  constructor(state, parent, action, g, h) {
    this.state = state; // for example, an object {x, y} representing the position in a maze
    this.parent = parent;
    this.action = action;
    this.g = g; // cost to reach this node
    this.h = h; // heuristic value of this node
    this.f = g + h; // priority of this node
  }`
        },
        {
          id: 'tab-3',
          title: 'Rust',
          lang: rust,
          code: `pub struct Node {
  pub state: State, // for example, a tuple (x, y) representing the position in a maze
  pub parent: Option<Box<Node>>,
  pub action: Option<Action>,
  pub g: i32, // cost to reach this node
  pub h: i32, // heuristic value of this node
  pub f: i32, // priority of this node
}

impl Node {
  pub fn new(state: State, parent: Option<Box<Node>>, action: Option<Action>, g: i32, h: i32) -> Node {
    Node {
      state,
      parent,
      action,
      g,
      h,
      f: g + h,
    }
  }
}`
        },
        {
          id: 'tab-4',
          title: 'C++',
          lang: cpp,
          code: `#include <optional>
#include <utility>

class Node {
public:
  State state; // for example, a std::pair<size_t, size_t> representing the position in a maze
  std::optional<Node*> parent;
  std::optional<Action> action;
  int g; // cost to reach this node
  int h; // heuristic value of this node
  int f; // priority of this node
};`
        }
      ]
    }/>
    <p class="my-4">
      In this example, we will focus on maze solving.
      Nodes represent a particular position in the maze and the links between nodes are
      the possible moves from one position to another (up, down, left, right).
      The initial node is the starting position and the goal is the position of the exit.
      The approach is to use some kind of frontier to store the possible paths to explore.
      At each step, the algorithm will move to a node in the frontier and refresh the
      frontier with the new possible paths. If the frontier is empty, it means we explored
      all the possible paths without finding the exit and the maze is unsolvable. Of course,
      the performance of the algorithm depends on the type of frontier used and the priorioty
      of the nodes to explore in the frontier.
    </p>
    <p class="mb-4">
      The code to solve the problem remains basically the same for all the search algorithms.
      Only the frontier changes.
    </p>
    <div class="mb-4 flex justify-center text-lg text-ctp-subtext1">
      <div>
        <p>Define the frontier</p>
        <p>Initialize the frontier with the initial node</p>
        <p>Define the explored set</p>
        <p>Loop indefinitely:</p>
        <p class="ml-4">If the frontier is empty:</p>
        <p class="ml-8"><i>There is no solution</i></p>
        <p class="ml-8">Return failure</p>
        <p class="ml-4">Let node be the node removed from the frontier</p>
        <p class="ml-4">If the node contains the goal state:</p>
        <p class="ml-8"><i>This is the solution</i></p>
        <p class="ml-8">Let solution be the path from the initial state to the node</p>
        <p class="ml-8">While the node has a parent:</p>
        <p class="ml-12">Add the action of the node to the beginning of the solution</p>
        <p class="ml-12">Let node be the parent of the node</p>
        <p class="ml-8">Return solution</p>
        <p class="ml-4">Else:</p>
        <p class="ml-8">Add all nodes that are not explored and reachable from the node to the frontier</p>
        <p class="ml-8">Add the node to the explored set</p>
      </div>
    </div>
  </section>
  <section class="mb-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">Depth-first search</h2>
    <p class="mb-2">
      Depth-first search (<strong>DFS</strong>) is a search algorithm that explores the graph by going as far as
      possible along each path before backtracking. It uses a stack frontier, which means
      that the last node added to the frontier will be the first one explored.
    </p>
    <p class="mb-2 pl-4 text-ctp-mauve">"Last in, first out"</p>
    <p class="mb-4">
      This algorithm is very simple to implement and is very efficient in terms of memory usage. However, it
      is not optimal in terms of performance. When a decision has to be made, it will choose
      the first possible path, which may not be the best one. Sometimes, it is lucky and find
      the exit quickly as you can see in maze <span class="bg-ctp-mauve rounded-md px-1.5 font-semibold text-ctp-mantle">4</span>.
      But in other cases, it will explore a lot of dead ends before finding the exit as you
      can see in maze <span class="bg-ctp-mauve rounded-md px-1.5 font-semibold text-ctp-mantle">2</span>.
      In addition, the algorithm will stop as soon as it finds a path to the exit, which means
      that it will not necessarily find the best one, as shown in maze
      <span class="bg-ctp-mauve rounded-md px-1.5 font-semibold text-ctp-mantle">1</span>.
    </p>
    <MultiLangCodeBlock code={
      [
        {
          id: 'tab-1',
          title: 'Python',
          lang: python,
          code: `class StackFontier():
  def __init__(self):
    self.frontier = []

  def add(self, node):
    self.frontier.append(node)

  def contains_state(self, state):
    return any(node.state == state for node in self.frontier)
  
  def is_empty(self):
    return len(self.frontier) == 0
  
  def remove(self):
    if self.is_empty():
      raise Exception("Empty Frontier")
    else:
      node = self.frontier[-1]
      self.frontier = self.frontier[:-1]
      return node
`
        },
        {
          id: 'tab-2',
          title: 'JavaScript',
          lang: javascript,
          code: `class StackFrontier {
  constructor() {
    this.frontier = [];
  }

  add(node) {
    this.frontier.push(node);
  }

  containsState(state) {
    return this.frontier.some((node) => node.state === state);
  }

  isEmpty() {
    return this.frontier.length === 0;
  }

  remove() {
    if (this.isEmpty()) {
      throw new Error("Empty Frontier");
    } else {
      return this.frontier.pop();
    }
  }
}`
        },
        {
          id: 'tab-3',
          title: 'Rust',
          lang: rust,
          code: `pub struct StackFrontier {
  frontier: Vec<Node>,
}

impl StackFrontier {
  pub fn new() -> StackFrontier {
    StackFrontier {
      frontier: Vec::new(),
    }
  }
  
  pub fn add(&mut self, node: Node) {
    self.frontier.push(node);
  }

  pub fn contains_state(&self, state: &State) -> bool {
    self.frontier.iter().any(|node| node.state == *state)
  }

  pub fn is_empty(&self) -> bool {
    self.frontier.is_empty()
  }

  pub fn remove(&mut self) -> Node {
    if self.is_empty() {
      panic!("Empty Frontier");
    } else {
      self.frontier.pop().unwrap()
    }
  }
}`
        },
        {
          id: 'tab-4',
          title: 'C++',
          lang: cpp,
          code: `#include <vector>
#include <algorithm>
#include <stdexcept>
          
class StackFrontier {
private:
  std::vector<Node> frontier;

public:
  StackFrontier() {}

  void add(Node node) {
    frontier.push_back(node);
  }

  bool containsState(State state) {
    return std::any_of(frontier.begin(), frontier.end(), [state](Node node) { return node.state == state; });
  }

  bool isEmpty() {
    return frontier.empty();
  }

  Node remove() {
    if (isEmpty()) {
      throw std::runtime_error("Empty Frontier");
    } else {
      Node node = frontier.back();
      frontier.pop_back();
      return node;
    }
  }
};`
        }
      ]
    }/>
  </section>
  <section class="my-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">Breadth-first search</h2>
    <p class="mb-2">
      Breadth-first search (<strong>BFS</strong>) explores all
      the nodes at the current depth before moving to the next depth. It uses a queue frontier,
      which means that the first node added to the frontier will be the first one explored.
    </p>
    <p class="mb-2 pl-4 text-ctp-mauve">"First in, first out"</p>
    <p class="mb-4">
      This algorithm is also very simple to implement and optimal in terms of performance as it will always find the shortest path
      to the exit. However, it is not optimal in terms of memory usage and execution time as it will explore all the possible paths
      at each depth before moving to the next one. This means that it will explore a lot of nodes that are not on the shortest path.
    </p>
    <MultiLangCodeBlock code={
      [
        {
          id: 'tab-1',
          title: 'Python',
          lang: python,
          code: `class QueueFontier():
  
  """ ... """
  
  def remove(self):
    if self.is_empty():
      raise Exception("Empty Frontier")
    else:
      node = self.frontier[0]
      self.frontier = self.frontier[1:]
      return node`
        },
        {
          id: 'tab-2',
          title: 'JavaScript',
          lang: javascript,
          code: `class QueueFrontier {
  
  /* ... */

  remove() {
    if (this.isEmpty()) {
      throw new Error("Empty Frontier");
    } else {
      return this.frontier.shift();
    }
  }
}`
        },
        {
          id: 'tab-3',
          title: 'Rust',
          lang: rust,
          code: `pub struct QueueFrontier {
  frontier: Vec<Node>,
}

impl QueueFrontier {

  /* ... */

  pub fn remove(&mut self) -> Node {
    if self.is_empty() {
      panic!("Empty Frontier");
    } else {
      self.frontier.remove(0)
    }
  }
}`
        },
        {
          id: 'tab-4',
          title: 'C++',
          lang: cpp,
          code: `class QueueFrontier {
  /* ... */

  Node remove() {
    if (isEmpty()) {
      throw std::runtime_error("Empty Frontier");
    } else {
      Node node = frontier.front();
      frontier.erase(frontier.begin());
      return node;
    }
  }
};`
        }
      ]
    }/>
  </section>
  <section class="mb-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">Uninformed vs informed search</h2>
    <p class="mb-4">
      DFS and BFS are examples of uninformed search algorithms. They do not use any information about the problem to guide the search.
    </p>
    <p class="mb-2">
      Informed search algorithms on the other hand use some information about the problem such as the distance to the goal or the cost of the path.
      This information is used to guide the search and explore the most promising paths first.
    </p>
    <p class="mb-2">
      Let's introduce the concept of heuristic. A heuristic is a function that estimates the distance from a node to the goal.
      This function often determines the efficiency of the algorithm. A good heuristic will guide the search to the goal quickly.
      For this example, we will use the Manhattan distance as a heuristic. The Manhattan distance is the sum of the horizontal and vertical
      distances between two points. It is the distance that a taxi would drive between two points in Manhattan.
    </p>
  </section>
  <section class="mb-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">Greedy best-first search</h2>
    <p class="mb-2">
      Greedy best-first search is an informed search algorithm that uses a priority queue frontier.
      The priority of a node is the heuristic value of the node. This means that the node with the lowest heuristic value will be the first one explored.
      This algorithm is quite efficient, however it will never go to a node that has a higher heuristic value than the current one unless it has no other choice.
      As you can see in maze <span class="bg-ctp-mauve rounded-md px-1.5 font-semibold text-ctp-mantle">4</span>, it doesn't find the shortest path because it
      doesn't take into account the cost of the path it is exploring.
    </p>
  </section>
  <section class="mb-8">
    <h2 class="mb-4 text-lg font-semibold text-ctp-lavender">A* search</h2>
    <p class="mb-2">
      A* search is also an informed search algorithm that uses a priority queue frontier.
      But, unlike greedy best-first, the priority of a node is not only the heuristic value of the node, but also the cost of the path to the node.
      This means that the node with the lowest sum of the heuristic value and the cost of the path will be the first one explored.
      In maze <span class="bg-ctp-mauve rounded-md px-1.5 font-semibold text-ctp-mantle">4</span>, where the greedy best-first failed to move to a node
      where the heuristic value was higher, the A* search will explore this node because the cost of the path is lower.
      This algorithm is very efficient and will always find the shortest path to the exit. It is the algorithm used today in most route planning applications.
    </p>
  </section>
  <section class="mb-8">
    <MultiLangCodeBlock code={
      [
        {
          id: 'tab-1',
          title: 'Python',
          lang: python,
          code: `class PriorityQueueFontier():
  
  """ ... """
  
  def add(self, node):
    self.frontier.append(node)
    self.frontier = sorted(self.frontier, key=lambda node: node.h) # or node.f for A* search

  def remove(self):
    if self.is_empty():
      raise Exception("Empty Frontier")
    else:
      node = self.frontier[0]
      self.frontier = self.frontier[1:]
      return node`
        },
        {
          id: 'tab-2',
          title: 'JavaScript',
          lang: javascript,
          code: `class PriorityQueueFontier {
  
  /* ... */

  add(node) {
    this.frontier.push(node);
    this.frontier.sort((a, b) => a.h - b.h); // or a.f - b.f for A* search
  }

  remove() {
    if (this.isEmpty()) {
      throw new Error("Empty Frontier");
    } else {
      return this.frontier.shift();
    }
  }
}`
        },
        {
          id: 'tab-3',
          title: 'Rust',
          lang: rust,
          code: `pub struct PriorityQueueFontier {
  frontier: Vec<Node>,
}

impl PriorityQueueFontier {

  /* ... */

  pub fn add(&mut self, node: Node) {
    self.frontier.push(node);
    self.frontier.sort_by(|a, b| a.h.cmp(&b.h)); // or a.f.cmp(&b.f) for A* search
  }

  pub fn remove(&mut self) -> Node {
    if self.is_empty() {
      panic!("Empty Frontier");
    } else {
      self.frontier.remove(0)
    }
  }
}`
        },
        {
          id: 'tab-4',
          title: 'C++',
          lang: cpp,
          code: `class PriorityQueueFontier {
  /* ... */

  void add(Node node) {
    frontier.push_back(node);
    std::sort(frontier.begin(), frontier.end(), [](Node a, Node b) { return a.h < b.h; }); // or a.f < b.f for A* search
  }

  Node remove() {
    if (isEmpty()) {
      throw std::runtime_error("Empty Frontier");
    } else {
      Node node = frontier.front();
      frontier.erase(frontier.begin());
      return node;
    }
  }
};`
        }
      ]
    }/>
  </section>
</div>
