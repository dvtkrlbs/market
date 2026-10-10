import { Heading } from "@astryxdesign/core/Heading";
import * as stylex from "@stylexjs/stylex";

const styles = stylex.create({
  heading: { padding: "var(--spacing-6)" },
});

export default function App() {
  return (
    <Heading level={1} xstyle={styles.heading}>
      Admin
    </Heading>
  );
}
