import { useEffect, useMemo, useRef, useState } from 'react'

function getOptionId(option) {
  return String(option?.id ?? option?.value ?? '')
}

function getOptionLabel(option) {
  return String(option?.name ?? option?.label ?? '')
}

export function ComboBox({
  id,
  value = '',
  options = [],
  placeholder = '',
  disabled = false,
  noResultsLabel = 'No results found',
  loadingLabel = 'Loading...',
  ariaBusy = false,
  onQueryChange,
  onChange,
}) {
  const rootRef = useRef(null)
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)

  const selectedOption = options.find((option) => getOptionId(option) === String(value))
  const selectedLabel = selectedOption ? getOptionLabel(selectedOption) : ''
  const filteredOptions = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase()
    if (!normalizedQuery) return options
    return options.filter((option) => getOptionLabel(option).toLowerCase().includes(normalizedQuery))
  }, [options, query])

  useEffect(() => {
    setQuery(selectedLabel)
    setActiveIndex(-1)
  }, [selectedLabel, value])

  useEffect(() => {
    const handleOutsidePointerDown = (event) => {
      if (rootRef.current?.contains(event.target)) return
      setOpen(false)
      setQuery(selectedLabel)
      setActiveIndex(-1)
    }

    document.addEventListener('pointerdown', handleOutsidePointerDown)
    return () => document.removeEventListener('pointerdown', handleOutsidePointerDown)
  }, [selectedLabel])

  const selectOption = (option) => {
    setQuery(getOptionLabel(option))
    setOpen(false)
    setActiveIndex(-1)
    onChange?.(getOptionId(option))
  }

  const resetQuery = () => {
    setQuery(selectedLabel)
    setOpen(false)
    setActiveIndex(-1)
  }

  const handleInputChange = (event) => {
    const nextQuery = event.target.value
    setQuery(nextQuery)
    setOpen(true)
    setActiveIndex(-1)
    onQueryChange?.(nextQuery)
    if (!nextQuery.trim()) onChange?.('')
  }

  const handleKeyDown = (event) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setOpen(true)
      setActiveIndex((current) => {
        if (filteredOptions.length === 0) return -1
        return current >= filteredOptions.length - 1 ? 0 : current + 1
      })
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      setOpen(true)
      setActiveIndex((current) => {
        if (filteredOptions.length === 0) return -1
        return current <= 0 ? filteredOptions.length - 1 : current - 1
      })
      return
    }

    if (event.key === 'Enter') {
      event.preventDefault()
      const activeOption = filteredOptions[activeIndex]
      const exactOption = filteredOptions.find(
        (option) => getOptionLabel(option).toLowerCase() === query.trim().toLowerCase(),
      )
      if (activeOption || exactOption) selectOption(activeOption || exactOption)
      return
    }

    if (event.key === 'Escape') {
      event.preventDefault()
      resetQuery()
    }
  }

  const handleBlur = () => {
    window.setTimeout(() => {
      if (!rootRef.current?.contains(document.activeElement)) resetQuery()
    }, 0)
  }

  return (
    <div ref={rootRef} className={`ks-combobox${open ? ' is-open' : ''}`}>
      <input
        id={id}
        className="form-control ks-combobox-input shadow-xs"
        value={query}
        placeholder={placeholder}
        disabled={disabled}
        role="combobox"
        aria-autocomplete="list"
        aria-controls={`${id}-listbox`}
        aria-expanded={open}
        aria-busy={ariaBusy}
        aria-haspopup="listbox"
        aria-activedescendant={activeIndex >= 0 ? `${id}-option-${activeIndex}` : undefined}
        onChange={handleInputChange}
        onFocus={() => !disabled && setOpen(true)}
        onKeyDown={handleKeyDown}
        onBlur={handleBlur}
      />
      <i className="bi bi-chevron-down ks-combobox-chevron" aria-hidden="true"></i>

      {open && !disabled && (
        <div id={`${id}-listbox`} className="ks-combobox-menu" role="listbox">
          {ariaBusy ? (
            <div className="ks-combobox-empty" role="status">
              {loadingLabel}
            </div>
          ) : filteredOptions.length === 0 ? (
            <div className="ks-combobox-empty" role="status">
              {noResultsLabel}
            </div>
          ) : (
            filteredOptions.map((option, index) => {
              const optionId = `${id}-option-${index}`
              const isSelected = getOptionId(option) === String(value)
              const isActive = index === activeIndex

              return (
                <div
                  id={optionId}
                  key={getOptionId(option)}
                  className={`ks-combobox-option${isActive ? ' is-active' : ''}`}
                  role="option"
                  aria-selected={isSelected}
                  onMouseDown={(event) => event.preventDefault()}
                  onMouseEnter={() => setActiveIndex(index)}
                  onClick={() => selectOption(option)}
                >
                  <span>{getOptionLabel(option)}</span>
                  {isSelected && <i className="bi bi-check2 text-primary" aria-hidden="true"></i>}
                </div>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}
