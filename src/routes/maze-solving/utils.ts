
type State = {
  x: number;
  y: number;
};

class Node {
  state: State;
  parent: Node | null;
  action: State | null;
  g: number; // cost
  h: number; // heuristic
  f: number; // g + h

  constructor(
    state: State,
    parent: Node | null,
    action: State | null,
    g: number, // cost
    h: number, // heuristic
  ) {
    this.state = state;
    this.parent = parent;
    this.action = action;
    this.g = g;
    this.h = h;
    this.f = g + h;
  };
}

class StackFrontier {
  frontier: Node[];

  constructor() {
    this.frontier = [];
  }

  add(node: Node) {
    this.frontier.push(node);
  }

  containsState(state: State) {
    return this.frontier.some(node => node.state.x === state.x && node.state.y === state.y);
  }

  isEmpty() {
    return this.frontier.length === 0;
  }

  remove() {
    if (this.isEmpty()) {
      throw new Error('Empty frontier');
    }
    return this.frontier.pop() as Node;
  }
}

class QueueFrontier extends StackFrontier {
  remove() {
    if (this.isEmpty()) {
      throw new Error('Empty frontier');
    }
    return this.frontier.shift() as Node;
  }
}

class PriorityQueueFrontier extends QueueFrontier {
  add(node: Node) {
    this.frontier.push(node);
    this.frontier.sort((a, b) => a.f - b.f);
  }

  remove() {
    if (this.isEmpty()) {
      throw new Error('Empty frontier');
    }
    return this.frontier.shift() as Node;
  }
}

export { type State, Node, StackFrontier, QueueFrontier, PriorityQueueFrontier };