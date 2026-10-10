import { createFileRoute } from "@tanstack/react-router";
import { RegisterPage } from "../RegisterPage";

export const Route = createFileRoute("/register")({ component: RegisterPage });
