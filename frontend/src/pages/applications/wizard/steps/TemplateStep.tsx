import {
  Tile,
  RadioButton,
  RadioButtonGroup,
  InlineLoading,
} from "@carbon/react";
import type { OnboardingTemplate } from "../../../../types/template";

type Props = {
  templates: OnboardingTemplate[];
  loading: boolean;
  selectedTemplateId?: string;
  onSelect: (template: OnboardingTemplate) => void;
};

export function TemplateStep({
  templates,
  loading,
  selectedTemplateId,
  onSelect,
}: Props) {
  if (loading) {
    return <InlineLoading description="Loading templates..." />;
  }

  return (
    <Tile>
      <h3>Select template</h3>
      <p>
        Choose the closest application type. The wizard will prefill defaults.
      </p>

      <RadioButtonGroup
        legendText="Templates"
        name="template"
        valueSelected={selectedTemplateId}
        onChange={(value) => {
          const selected = templates.find((item) => item.id === value);
          if (selected) onSelect(selected);
        }}
      >
        {templates.map((template) => (
          <RadioButton
            key={template.id}
            id={template.id}
            labelText={`${template.name} — ${template.description}`}
            value={template.id}
          />
        ))}
      </RadioButtonGroup>
    </Tile>
  );
}
