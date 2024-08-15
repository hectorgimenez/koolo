package step

import (
    "fmt"
    "time"
    "math/rand"
    "github.com/hectorgimenez/koolo/internal/container"
    "github.com/hectorgimenez/koolo/internal/game"
    "github.com/hectorgimenez/d2go/pkg/data"
    "github.com/hectorgimenez/d2go/pkg/data/object"
    "github.com/hectorgimenez/koolo/internal/helper"
    "github.com/hectorgimenez/koolo/internal/pather"
)

type InteractObjectStep struct {
    basicStep
    objectName            object.Name
    objectID              data.UnitID
    waitingForInteraction bool
    isCompleted           func(game.Data) bool
    mouseOverAttempts     int
    currentMouseCoords    data.Position
    startTime             time.Time
    teleportAttempts      int
}

func InteractObject(name object.Name, isCompleted func(game.Data) bool) *InteractObjectStep {
    return &InteractObjectStep{
        basicStep:   newBasicStep(),
        objectName:  name,
        isCompleted: isCompleted,
        startTime:   time.Now(),
    }
}

func InteractObjectByID(ID data.UnitID, isCompleted func(game.Data) bool) *InteractObjectStep {
    return &InteractObjectStep{
        basicStep:   newBasicStep(),
        objectID:    ID,
        isCompleted: isCompleted,
        startTime:   time.Now(),
    }
}

func (i *InteractObjectStep) Status(d game.Data, _ container.Container) Status {
    if i.status == StatusCompleted {
        return StatusCompleted
    }
    if i.isCompleted != nil && i.isCompleted(d) {
        return i.tryTransitionStatus(StatusCompleted)
    }
    return i.status
}

func (i *InteractObjectStep) Run(d game.Data, container container.Container) error {
    i.tryTransitionStatus(StatusInProgress)

    // Check if the interaction is already completed
    if i.isCompleted != nil && i.isCompleted(d) {
        i.tryTransitionStatus(StatusCompleted)
        return nil
    }

    // Check for timeout
    const timeout = 8 * time.Second
    const maxTeleportAttempts = 5

    if time.Since(i.startTime) > timeout {
        if i.teleportAttempts >= maxTeleportAttempts {
            return fmt.Errorf("max teleport attempts reached for object %d", i.objectName)
        }

        // Timeout exceeded, teleport to a random position near the object
        var o data.Object
        var found bool
        if i.objectID != 0 {
            for _, obj := range d.Objects {
                if obj.ID == i.objectID {
                    o = obj
                    found = true
                    break
                }
            }
        } else {
            o, found = d.Objects.FindOne(i.objectName)
        }

        if !found {
            return fmt.Errorf("object %d not found", i.objectName)
        }

        randX := rand.Intn(11) - 5
        randY := rand.Intn(11) - 5
        newPos := data.Position{X: o.Position.X + randX, Y: o.Position.Y + randY}

        moveStep := MoveTo(newPos)
        err := moveStep.Run(d, container)
        if err != nil {
            return fmt.Errorf("failed to teleport: %v", err)
        }

        i.teleportAttempts++
        i.startTime = time.Now()
        i.mouseOverAttempts = 0
        i.waitingForInteraction = false
        return nil
    }

    if i.mouseOverAttempts > maxInteractions {
        return fmt.Errorf("object %d could not be interacted", i.objectName)
    }

    if i.waitingForInteraction && time.Since(i.lastRun) < time.Millisecond*500 {
        return nil
    }

    i.lastRun = time.Now()

    var o data.Object
    var found bool
    if i.objectID != 0 {
        for _, obj := range d.Objects {
            if obj.ID == i.objectID {
                o = obj
                found = true
                break
            }
        }
    } else {
        o, found = d.Objects.FindOne(i.objectName)
        if i.objectName == object.TownPortal {
            for _, obj := range d.Objects {
                if obj.Owner == d.PlayerUnit.Name {
                    o = obj
                    found = true
                }
            }
        }
    }

    if found {
        objectX := o.Position.X - 2
        objectY := o.Position.Y - 2
        if o.IsHovered {
            container.HID.Click(game.LeftButton, i.currentMouseCoords.X, i.currentMouseCoords.Y)
            i.waitingForInteraction = true
            return nil
        } else {
            distance := pather.DistanceFromMe(d, o.Position)
            if distance > 15 {
                return fmt.Errorf("object is too far away: %d. Current distance: %d", o.Name, distance)
            }
            mX, mY := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX, objectY)
            if i.mouseOverAttempts == 2 && o.Name == object.TownPortal {
                mX, mY = container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, objectX-4, objectY-4)
            }
            x, y := helper.Spiral(i.mouseOverAttempts)
            i.currentMouseCoords = data.Position{X: mX + x, Y: mY + y}
            container.HID.MovePointer(mX+x, mY+y)
            i.mouseOverAttempts++
            return nil
        }
    }

    return fmt.Errorf("object %d not found", i.objectName)
}

func (i *InteractObjectStep) Reset() {
    i.basicStep.Reset()
    i.mouseOverAttempts = 0
    i.waitingForInteraction = false
    i.startTime = time.Now()
    i.teleportAttempts = 0
}
