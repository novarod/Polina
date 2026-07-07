import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it } from "vitest";

import { InputField } from "@/components/form/input-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";

function EmailForm() {
  const form = useForm<{ email: string }>({ defaultValues: { email: "" } });
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(() => {})} noValidate>
        <InputField
          control={form.control}
          name="email"
          label="Email"
          type="email"
        />
        <Button type="submit">Enviar</Button>
      </form>
    </Form>
  );
}

describe("InputField", () => {
  it("binds the label to the input and forwards value changes", async () => {
    const user = userEvent.setup();
    render(<EmailForm />);

    const input = screen.getByLabelText("Email");
    await user.type(input, "a@b.com");

    expect(input).toHaveValue("a@b.com");
  });
});
