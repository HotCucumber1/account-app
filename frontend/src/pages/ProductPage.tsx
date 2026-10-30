import {Link} from 'react-router-dom'

export function ProductPage() {
    return (
        <section className="card center">
            <h1>Продукт</h1>
            <img src="/cat.gif" alt="Анимация кота" width={240} height={240}/>
            <p>
                <Link to="/">Вернуться в личный кабинет</Link>
            </p>
        </section>
    )
}
