import { createFileRoute } from "@tanstack/react-router";
import { StoresPage } from "../StoresPage";

export const Route = createFileRoute("/stores")({ component: StoresPage });
