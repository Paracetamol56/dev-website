import { Node, type Frontier, type State } from "./utils";

class Cell {
  wall: boolean;
  start: boolean;
  end: boolean;
  visited: boolean;
  solution: boolean;

  constructor(wall: boolean) {
    this.wall = wall;
    this.start = false;
    this.end = false;
    this.visited = false;
    this.solution = false;
  }
}

class Maze<F extends Frontier> {
  height: number;
  width: number;
  cells: Cell[][] = [];
  start: State = { x: 0, y: 0 };
  end: State = { x: 0, y: 0 };
  lastNode: Node | null = null;
  explored: Node[] = [];
  solved: boolean = false;
  solvable: boolean = true;
  frontier: F;
  stepCallback: (node: Node) => void;

  /**
   * Creates a new instance of the Maze class.
   * @param frontier The type of frontier to use for the maze solving algorithm.
   * @param stepCallback Optional callback function to be called on each step of the maze solving algorithm.
   */
  constructor(frontier: F, stepCallback: (node: Node) => void) {
    this.frontier = frontier;
    this.stepCallback = stepCallback;
    this.height = 0;
    this.width = 0;
  }

  /**
   * Loads a maze from a string representation.
   * @param mazeString - The string representation of the maze.
   */
  loadFromString(mazeString: string) {
    const splitedMazeString = mazeString.split("\n");
    this.height = splitedMazeString.length;
    this.width = splitedMazeString[0].length;

    this.cells = new Array(this.height);
    for (let i = 0; i < this.height; i++) {
      this.cells[i] = new Array(this.width);
      for (let j = 0; j < this.width; j++) {
        this.cells[i][j] = new Cell(splitedMazeString[i][j] === "#");
        if (splitedMazeString[i][j] === "A") {
          this.start = { x: i, y: j};
          this.cells[i][j].start = true;
        } else if (splitedMazeString[i][j] === "B") {
          this.end = { x: i, y: j};
          this.cells[i][j].end = true;
        }
      }
    }
    this.solved = false;
    this.solvable = true;
    this.explored = [];
    this.frontier.add(
      new Node(
        this.start,
        null,
        null,
        0,
        this.heuristic(this.start, this.end)
      )
    );
    // First step
    this.lastNode = this.move();
  }

  reset() {
    this.solved = false;
    this.solvable = true;
    this.explored = [];
    this.frontier.clear();
    this.frontier.add(
      new Node(
        this.start,
        null,
        null,
        0,
        this.heuristic(this.start, this.end)
      )
    );
    for (const row of this.cells) {
      for (const cell of row) {
        cell.visited = false;
        cell.solution = false;
      }
    }
    this.lastNode = this.move();
    this.stepCallback(this.lastNode);
  }

  /**
   * Returns the manhattan distance between two points.
   * @param a The first point.
   * @param b The second point.
   * @returns The manhattan distance between the two points.
   */
  heuristic(a: State, b: State) {
    return Math.abs(a.x - b.x) + Math.abs(a.y - b.y);
  }

  /**
   * Returns the list of neighbors of a given cell.
   * @param state The cell to get the neighbors of.
   * @returns The list of neighbors of the given cell.
   */
  neighbors(state: State) {
    const { x, y } = state;
    const candidates = [
      [x + 1, y],
      [x - 1, y],
      [x, y + 1],
      [x, y - 1],
    ];
    const neighbors: State[] = [];
    for (const [x, y] of candidates) {
      if (0 <= x && x < this.height && 0 <= y && y < this.width && !this.cells[x][y].wall) {
        neighbors.push({x, y});
      }
    }

    return neighbors;
  }

  /**
   * Moves the current node to the next node in the frontier.
   * @returns The next node in the frontier.
   * @throws An error if there is no solution.
   */
  move(): Node {
    if (this.frontier.isEmpty()) {
      throw new Error("No solution");
    }

    const node = this.frontier.remove();
    this.explored.push(node);
    this.cells[node!.state.x][node!.state.y].visited = true;

    for (const neighbor of this.neighbors(node!.state)) {
      const child = new Node(
        neighbor,
        node,
        null,
        node!.g + 1,
        this.heuristic(neighbor, this.end)
      );

      if (
        !this.frontier.containsState(child.state) &&
        !this.explored.some((node) => node.state.x === child.state.x && node.state.y === child.state.y)
      ) {
        this.frontier.add(child);
      }
    }

    return node;
  }

  /**
   * Returns the path to the given end node.
   * @param endNode - The end node to find the path to.
   * @returns An array of States representing the path to the end node.
   */
  getPathTo(endNode: Node) {
    const solvedPath: State[] = [];
    let node: Node = endNode;
    while (node.parent) {
      solvedPath.push(node.state);
      this.cells[node.state.x][node.state.y].solution = true;
      node = node.parent;
    }
    solvedPath.push(node.state);
    this.cells[node.state.x][node.state.y].solution = true;
    return solvedPath;
  }

  /**
   * Performs a single step of the AI algorithm to solve the maze.
   * If the maze is already solved, this function does nothing.
   * 
   * @returns The last node visited during the step, or null if no node was visited.
   * @throws An error if the maze has no solution.
   */
  aiStep() {
    if (this.solved) return;

    try {
      this.lastNode = this.move();
      if (this.lastNode?.state.x === this.end.x && this.lastNode?.state.y === this.end.y) {
        this.solved = true;
        for (const state of this.getPathTo(this.lastNode)) {
          this.cells[state.x][state.y].solution = true;
        }
      }
      this.stepCallback(this.lastNode);
    } catch (e) {
      console.log(e);
      if (e instanceof Error) {
        if (e.message == "No solution") {
          this.solvable = false;
        } else {
          console.error(e);
        }
      }
    }
  }

  /**
   * Solves the maze using an AI algorithm.
   * @returns void
   */
  aiSolve() {
    if (this.solved) return;

    try {
      while (true) {
        this.lastNode = this.move();
        if (this.lastNode?.state.x == this.end.x && this.lastNode?.state.y == this.end.y) {
          this.solved = true;
          for (const state of this.getPathTo(this.lastNode)) {
            this.cells[state.x][state.y].solution = true;
          }
          this.stepCallback(this.lastNode);
          break;
        }
        this.stepCallback(this.lastNode);
      }
    } catch (e) {
      console.log(e);
      if (e instanceof Error) {
        if (e.message == "No solution") {
          this.solvable = false;
        } else {
          console.error(e);
        }
      }
    }
  }
};

export { Cell, Maze };
