import { writable, type Writable } from 'svelte/store';

// Abstract base class for debounced operations
export abstract class DebouncedOperation<TArgs extends any[], TResult> {
  protected timeoutId: NodeJS.Timeout | null = null;
  protected lastArgs: TArgs | null = null;
  protected hasChanged = false;
  protected _store: Writable<TResult | null> | null = null;

  constructor(
    protected delay: number = 1000,
    protected options: {
      useStore?: boolean;
      onError?: (error: any) => void;
      onSuccess?: (result: TResult) => void;
    } = {}
  ) {
    if (options.useStore) {
      this._store = writable(null);
    }
  }

  // Abstract method that subclasses must implement
  protected abstract executeOperation(...args: TArgs): TResult | Promise<TResult>;

  // Abstract method for comparing arguments (can be overridden)
  protected compareArgs(prev: TArgs, current: TArgs): boolean {
    if (prev.length !== current.length) return false;
    return prev.every((val, index) => val === current[index]);
  }

  // Main execution method
  private async performExecution(...args: TArgs): Promise<TResult | void> {
    try {
      const result = await this.executeOperation(...args);

      // Update store if using one
      if (this._store) {
        this._store.set(result);
      }

      // Call success handler if provided
      if (this.options.onSuccess) {
        this.options.onSuccess(result);
      }

      this.lastArgs = args;
      this.hasChanged = false;

      return result;
    } catch (error) {
      this.handleError(error);
      throw error;
    }
  }

  // Error handling (can be overridden)
  protected handleError(error: any): void {
    if (this.options.onError) {
      this.options.onError(error);
    } else {
      console.error(`${this.constructor.name} error:`, error);
    }
  }

  // Debounced execution
  public execute(...args: TArgs): void {
    // Check if arguments have changed
    if (!this.lastArgs || !this.compareArgs(this.lastArgs, args)) {
      this.hasChanged = true;
    }

    // Clear existing timeout
    this.clearTimeout();

    // Set new timeout
    this.timeoutId = setTimeout(() => {
      if (this.hasChanged) {
        this.performExecution(...args);
      }
    }, this.delay);
  }

  // Immediate execution (bypass debouncing)
  public async executeNow(...args: TArgs): Promise<TResult | void> {
    this.clearTimeout();
    return this.performExecution(...args);
  }

  // Cancel pending execution
  public cancel(): void {
    this.clearTimeout();
    this.hasChanged = false;
  }

  // Cleanup resources
  public cleanup(): void {
    this.clearTimeout();
    this.lastArgs = null;
    this.hasChanged = false;
  }

  // Get the store (if created)
  public get store(): Writable<TResult | null> | null {
    return this._store;
  }

  // Check if execution is pending
  public get isPending(): boolean {
    return this.timeoutId !== null;
  }

  // Update delay
  public setDelay(newDelay: number): void {
    this.delay = newDelay;
  }

  // Private helper to clear timeout
  private clearTimeout(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
  }
}

