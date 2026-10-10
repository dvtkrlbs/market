import { Heading } from "@astryxdesign/core/Heading";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: () => <Heading level={1}>Customer</Heading>,
});
