package mocks

import "github.com/sclevine/agouti/core/internal/types"

type Element struct {
	GetIDCall struct {
		ReturnID string
	}

	GetElementCall struct {
		Selector      types.Selector
		ReturnElement types.Element
		Err           error
	}

	GetElementsCall struct {
		Selector       types.Selector
		ReturnElements []types.Element
		Err            error
	}

	GetTextCall struct {
		ReturnText string
		Err        error
	}

	GetAttributeCall struct {
		Attribute   string
		ReturnValue string
		Err         error
	}

	GetCSSCall struct {
		Property    string
		ReturnValue string
		Err         error
	}

	ClickCall struct {
		Called bool
		Err    error
	}

	ClearCall struct {
		Called bool
		Err    error
	}

	ValueCall struct {
		Text string
		Err  error
	}

	IsSelectedCall struct {
		ReturnSelected bool
		Err            error
	}

	IsDisplayedCall struct {
		ReturnDisplayed bool
		Err             error
	}

	IsEnabledCall struct {
		ReturnEnabled bool
		Err           error
	}

	SubmitCall struct {
		Called bool
		Err    error
	}

	IsEqualToCall struct {
		Element      types.Element
		ReturnEquals bool
		Err          error
	}
}

func (e *Element) GetID() string {
	return e.GetIDCall.ReturnID
}

func (e *Element) GetElement(selector types.Selector) (types.Element, error) {
	e.GetElementCall.Selector = selector
	return e.GetElementCall.ReturnElement, e.GetElementCall.Err
}

func (e *Element) GetElements(selector types.Selector) ([]types.Element, error) {
	e.GetElementsCall.Selector = selector
	return e.GetElementsCall.ReturnElements, e.GetElementsCall.Err
}

func (e *Element) GetText() (string, error) {
	return e.GetTextCall.ReturnText, e.GetTextCall.Err
}

func (e *Element) GetAttribute(attribute string) (string, error) {
	e.GetAttributeCall.Attribute = attribute
	return e.GetAttributeCall.ReturnValue, e.GetAttributeCall.Err
}

func (e *Element) GetCSS(property string) (string, error) {
	e.GetCSSCall.Property = property
	return e.GetCSSCall.ReturnValue, e.GetCSSCall.Err
}

func (e *Element) Click() error {
	e.ClickCall.Called = true
	return e.ClickCall.Err
}

func (e *Element) Clear() error {
	e.ClearCall.Called = true
	return e.ClearCall.Err
}

func (e *Element) Value(text string) error {
	e.ValueCall.Text = text
	return e.ValueCall.Err
}

func (e *Element) IsSelected() (bool, error) {
	return e.IsSelectedCall.ReturnSelected, e.IsSelectedCall.Err
}

func (e *Element) IsDisplayed() (bool, error) {
	return e.IsDisplayedCall.ReturnDisplayed, e.IsDisplayedCall.Err
}

func (e *Element) IsEnabled() (bool, error) {
	return e.IsEnabledCall.ReturnEnabled, e.IsEnabledCall.Err
}

func (e *Element) Submit() error {
	e.SubmitCall.Called = true
	return e.SubmitCall.Err
}

func (e *Element) IsEqualTo(other types.Element) (bool, error) {
	e.IsEqualToCall.Element = other
	return e.IsEqualToCall.ReturnEquals, e.IsEqualToCall.Err
}
