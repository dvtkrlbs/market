import { createFileRoute } from "@tanstack/react-router";
import { ApplyPage } from "../ApplyPage";

export const Route = createFileRoute("/apply")({ component: ApplyPage });
