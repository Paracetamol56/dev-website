import { Node, PriorityQueueFrontier, QueueFrontier, StackFrontier, type State } from "./utils";

class Cell {
  wall: boolean;
  start: boolean;
  end: boolean;
  visited: boolean;
  solution: boolean;
  element: HTMLDivElement | null = null;

  constructor(wall: boolean) {
    this.wall = wall;
    this.start = false;
    this.end = false;
    this.visited = false;
    this.solution = false;
  }
}

class Maze {
  height: number;
  width: number;
  cells: Cell[][] = [];
  start: State = { x: 0, y: 0 };
  end: State = { x: 0, y: 0 };
  explored: State[] = [];
  frontier: PriorityQueueFrontier = new PriorityQueueFrontier();

  /**
   * Creates a new maze.
   * @param mazeString The string representation of the maze.
   *   - `#` represents a wall.
   *   - ` ` represents an empty cell.
   *   - `A` represents the start cell.
   *   - `B` represents the end cell.
   *   - `/` represents a new line.
   */
  constructor(mazeString: string) {
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
    this.frontier.add(
      new Node(
        this.start,
        null,
        null,
        0,
        this.heuristic(this.start, this.end)
      )
    );
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

  move(): Node {
    if (this.frontier.isEmpty()) {
      throw new Error("No solution");
    }

    const node = this.frontier.remove();
    this.explored.push(node!.state);
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
        !this.explored.some((state) => state.x === child.state.x && state.y === child.state.y)
      ) {
        this.frontier.add(child);
      }
    }

    return node;
  }
};

export { Cell, Maze };