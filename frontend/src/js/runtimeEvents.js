export const runtimeChangedEvent = "wip:runtime-changed";

export function notifyRuntimeChanged() {
  window.dispatchEvent(new Event(runtimeChangedEvent));
}